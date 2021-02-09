package ttl

import (
	"container/list"
	"sync"
	"time"
)

// Cache is a synchronised map of items that auto-expire once stale
type Cache struct {
	mutex    sync.RWMutex
	capacity int
	ttl      time.Duration
	items    map[interface{}]Item
	lruList  *list.List // realize LikedHashMap with items
	ttlList  *list.List // realize LikedHashMap with items
	timer    *time.Timer
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

// NewCache create a cache which maintains permanence k-v
func NewCache() *Cache {
	c := &Cache{
		items:    make(map[interface{}]Item),
		shutdown: make(chan struct{}),
	}
	return c
}

func NewCacheWithLRU(cap int) *Cache {
	c := &Cache{
		items:    make(map[interface{}]Item),
		shutdown: make(chan struct{}),
	}
	if cap > 0 {
		c.capacity = cap
		c.lruList = list.New()
	}
	return c
}

func NewCacheWithTTL(d time.Duration) *Cache {
	c := &Cache{
		items:    make(map[interface{}]Item),
		shutdown: make(chan struct{}),
	}
	c.ttl = d
	c.ttlList = list.New()
	return c
}

// Set is a thread-safe way to add new items to the map
func (c *Cache) Set(key, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
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
	}
	if c.lruList != nil {
		item.lru = c.lruList.PushFront(&ent)
	}
	// add new
	c.items[key] = item
	// timer
	if c.timer == nil {
		go c.startCleanupTimer()
	}

	// remove least used if over maximum capacity
	if c.capacity > 0 && len(c.items) > c.capacity {
		if e := c.lruList.Back(); e != nil {
			c.remove(e.Value.(*entry).key)
		}
	}
}

// Get is a thread-safe way to lookup items.
// Every lookup, also hence extending it's life if ttl is enabled
func (c *Cache) Get(key interface{}) (interface{}, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	item, ok := c.items[key]
	if !ok {
		return nil, false
	}

	if item.expireAt.After(time.Now()) {
		if c.capacity > 0 {
			c.lruList.MoveToFront(item.lru)
		}
		return item.value, true
	}
	// expired
	c.remove(key)
	return nil, false
}

func (c *Cache) Remove(key interface{}) {
	c.mutex.Lock()
	c.remove(key)
	c.mutex.Unlock()
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
	c.mutex.RLock()
	lens := len(c.items)
	c.mutex.RUnlock()
	return lens
}

func (c *Cache) Close() {
	c.mutex.Lock()
	close(c.shutdown)
	c.items = nil
	c.lruList = nil
	c.ttlList = nil
	c.mutex.Unlock()
}

// cleanup cleanup expired items and reset timer for next cleanup
func (c *Cache) cleanup() {
	for {
		if e := c.ttlList.Back(); e != nil {
			key := e.Value.(*entry).key
			d := time.Until(c.items[key].expireAt)
			if d <= 0 {
				c.remove(key)
				continue
			} else if d < time.Second {
				d = time.Second
			}
			// seems not necessary to stop
			// if !c.timer.Stop() {
			// 	select {
			// 	case <-c.timer.C: // try to drain from the channel
			// 	default:
			// 	}
			// }
			c.timer.Reset(d)
		}
		return
	}
}

// TODO: timewheel
// https://github.com/rfyiamcool/golib/blob/29fd190076471bbc97809eba72c25cacd51dccf5/timewheel/tw.go
func (c *Cache) startCleanupTimer() {
	c.timer = time.NewTimer(c.ttl)
	for {
		select {
		case <-c.shutdown:
			return
		case <-c.timer.C:
			c.mutex.Lock()
			c.cleanup()
			c.mutex.Unlock()
		}
	}
}
