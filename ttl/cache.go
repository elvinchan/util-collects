package ttl

import (
	"container/list"
	"sync"
	"time"
)

var defaultMinGap = time.Second

// Cache is a synchronised map of items that auto-expire once stale
type Cache struct {
	sync.RWMutex
	cap      int
	ttl      time.Duration
	items    map[interface{}]Item
	lruList  *list.List // realize LikedHashMap with items
	ttlList  *list.List // realize LikedHashMap with items
	timer    *time.Timer
	cleaning bool
	shutdown chan struct{}
}

type Item struct {
	value    interface{}
	expireAt time.Time
	lru      *list.Element
	ttl      *list.Element
}

type entry struct {
	key interface{} // 如果key比较大，会造成内存占用高
}

type CacheOption func(*Cache)

func CacheWithLRU(cap int) CacheOption {
	return func(c *Cache) {
		if cap > 0 {
			c.cap = cap
			c.lruList = list.New()
		}
	}
}

func CacheWithTTL(d time.Duration) CacheOption {
	return func(c *Cache) {
		if d < defaultMinGap {
			d = defaultMinGap
		}
		c.ttl = d
		c.ttlList = list.New()
	}
}

// NewCache create a K-V cache
func NewCache(opts ...CacheOption) *Cache {
	c := &Cache{
		items:    make(map[interface{}]Item),
		shutdown: make(chan struct{}),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Set is a thread-safe way to set item to the map
func (c *Cache) Set(key, value interface{}) {
	c.Lock()
	defer c.Unlock()
	// already exist
	if item, ok := c.items[key]; ok {
		if c.lruList != nil {
			c.lruList.MoveToFront(item.lru)
		}
		item.value = value
		return
	}

	item := Item{
		value: value,
	}
	ent := entry{
		key,
	}
	if c.ttlList != nil {
		item.expireAt = time.Now().Add(c.ttl)
		item.ttl = c.ttlList.PushFront(&ent)
		if !c.cleaning {
			if c.timer == nil {
				c.timer = time.NewTimer(c.ttl)
			} else {
				// reset timer
				c.timer.Reset(c.ttl)
			}
			c.cleaning = true
			go c.startCleanup()
		}
	}
	if c.lruList != nil {
		item.lru = c.lruList.PushFront(&ent)
	}
	// add new
	c.items[key] = item

	// remove least used if over maximum capacity
	if c.cap > 0 && len(c.items) > c.cap {
		if e := c.lruList.Back(); e != nil {
			c.remove(e.Value.(*entry).key)
		}
	}
}

// Get is a thread-safe way to lookup items.
// Every lookup, also hence extending it's life if ttl is enabled
func (c *Cache) Get(key interface{}) (interface{}, bool) {
	c.Lock()
	defer c.Unlock()

	item, ok := c.items[key]
	if !ok {
		return nil, false
	}

	if item.expireAt.IsZero() || item.expireAt.After(time.Now()) {
		if c.cap > 0 {
			c.lruList.MoveToFront(item.lru)
		}
		return item.value, true
	}
	// expired
	c.remove(key)
	return nil, false
}

// Remove removes item from Cache
func (c *Cache) Remove(key interface{}) {
	c.Lock()
	c.remove(key)
	c.Unlock()
}

func (c *Cache) remove(key interface{}) {
	if c.items[key].lru != nil {
		c.lruList.Remove(c.items[key].lru)
	}
	if c.items[key].ttl != nil {
		c.ttlList.Remove(c.items[key].ttl)
	}
	delete(c.items, key)
}

// Len returns the number of items in the cache
func (c *Cache) Len() int {
	c.RLock()
	lens := len(c.items)
	c.RUnlock()
	return lens
}

// Reset remove all data from Cache so it can use again
func (c *Cache) Reset() {
	if !c.timer.Stop() {
		select {
		case <-c.timer.C: // try to drain from the channel
		default:
		}
	}
	c.Lock()
	c.items = make(map[interface{}]Item)
	if c.lruList != nil {
		c.lruList.Init()
	}
	if c.ttlList != nil {
		c.ttlList.Init()
	}
	c.Unlock()
}

// Close remove all data from Cache and exit cleanup.
// Cache cannot use any more after close
func (c *Cache) Close() {
	c.Lock()
	close(c.shutdown)
	c.items = nil
	c.lruList = nil
	c.ttlList = nil
	c.cleaning = false
	c.Unlock()
}

// cleanup cleanup expired items and reset timer for next cleanup
// if there's no more data in ttlList, exit
func (c *Cache) cleanup() bool {
	for {
		if e := c.ttlList.Back(); e != nil {
			key := e.Value.(*entry).key
			d := time.Until(c.items[key].expireAt)
			if d <= 0 {
				c.remove(key)
				continue
			} else if d < defaultMinGap {
				d = defaultMinGap
			}
			// seems not necessary to stop
			// if !c.timer.Stop() {
			// 	select {
			// 	case <-c.timer.C: // try to drain from the channel
			// 	default:
			// 	}
			// }
			c.timer.Reset(d)
			return false
		}
		return true
	}
}

// TODO: timewheel
// https://github.com/rfyiamcool/golib/blob/29fd190076471bbc97809eba72c25cacd51dccf5/timewheel/tw.go
func (c *Cache) startCleanup() {
	for {
		select {
		case <-c.shutdown:
			c.timer.Stop()
			return
		case <-c.timer.C:
			c.Lock()
			if c.cleanup() {
				c.Unlock()
				c.cleaning = false
				return
			}
			c.Unlock()
		}
	}
}
