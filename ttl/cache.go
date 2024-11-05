package ttl

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"
)

var defaultMinGap = time.Second

// Cache is a synchronised map of items that auto-expire once stale
type Cache struct {
	mu             sync.RWMutex
	cap            int
	ttl            time.Duration
	ttlRefreshMode TTLRefreshMode
	items          map[interface{}]*Item
	lruList        *list.List // realize LinkedHashMap with items
	ttlList        *list.List // realize LinkedHashMap with items
	timer          *time.Timer
	cleaning       uint32 // 0 -> false, 1 -> true
	shutdown       chan struct{}
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

type TTLRefreshMode uint8

const (
	TTLRefreshModeGet TTLRefreshMode = 1 << iota
	TTLRefreshModeSet
	TTLRefreshModeNone TTLRefreshMode = 0
)

func CacheWithTTL(d time.Duration, refreshMode TTLRefreshMode) CacheOption {
	return func(c *Cache) {
		if d < defaultMinGap {
			d = defaultMinGap
		}
		c.ttl = d
		c.ttlRefreshMode = refreshMode
		c.ttlList = list.New()
	}
}

// NewCache create a K-V cache
func NewCache(opts ...CacheOption) *Cache {
	c := &Cache{
		items:    make(map[interface{}]*Item),
		shutdown: make(chan struct{}),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Set is a thread-safe way to set item to the map
func (c *Cache) Set(key, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if item, ok := c.items[key]; ok {
		// already exist
		if c.lruList != nil {
			c.lruList.MoveToFront(item.lru)
		}
		if c.ttlList != nil && c.ttlRefreshMode&TTLRefreshModeSet > 0 {
			item.expireAt = time.Now().Add(c.ttl)
			c.ttlList.MoveToFront(item.ttl)
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
	if c.lruList != nil {
		item.lru = c.lruList.PushFront(&ent)
	}
	if c.ttlList != nil {
		item.expireAt = time.Now().Add(c.ttl)
		item.ttl = c.ttlList.PushFront(&ent)
		if atomic.CompareAndSwapUint32(&c.cleaning, 0, 1) {
			go c.startCleanup()
		}
	}
	// add new
	c.items[key] = &item

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
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.items[key]
	if !ok {
		return nil, false
	}

	if item.expireAt.IsZero() || item.expireAt.After(time.Now()) {
		if c.lruList != nil {
			c.lruList.MoveToFront(item.lru)
		}
		if c.ttlList != nil && c.ttlRefreshMode&TTLRefreshModeGet > 0 {
			item.expireAt = time.Now().Add(c.ttl)
			c.ttlList.MoveToFront(item.ttl)
		}
		return item.value, true
	}
	// expired
	c.remove(key)
	return nil, false
}

// Remove removes item from Cache
func (c *Cache) Remove(key interface{}) {
	c.mu.Lock()
	c.remove(key)
	c.mu.Unlock()
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
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Close remove all data from Cache and exit cleanup.
// Cache cannot use any more after close
func (c *Cache) Close() {
	close(c.shutdown)
	c.mu.Lock()
	c.items = nil
	c.lruList = nil
	c.ttlList = nil
	c.mu.Unlock()
}

func (c *Cache) ttlNextExpire() (bool, time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ttlList == nil {
		return true, 0
	}
	if e := c.ttlList.Back(); e != nil {
		key := e.Value.(*entry).key
		d := time.Until(c.items[key].expireAt)
		if d < 0 {
			c.remove(key)
		}
		return false, d
	}
	return true, 0
}

// cleanup cleanup expired items and reset timer for next cleanup
// if there's no more data in ttlList, exit
func (c *Cache) cleanup() bool {
	for {
		empty, d := c.ttlNextExpire()
		if empty {
			return true
		}
		if d <= 0 {
			continue
		} else if d < defaultMinGap {
			d = defaultMinGap
		}
		c.timer.Reset(d)
		return false
	}
}

// TODO: timewheel
// https://github.com/rfyiamcool/golib/blob/29fd190076471bbc97809eba72c25cacd51dccf5/timewheel/tw.go
func (c *Cache) startCleanup() {
	if c.timer == nil {
		c.timer = time.NewTimer(c.ttl)
	} else {
		c.timer.Reset(c.ttl)
	}
	for {
		select {
		case <-c.shutdown:
			c.timer.Stop()
			return
		case <-c.timer.C:
			if c.cleanup() {
				atomic.StoreUint32(&c.cleaning, 0)
				return
			}
		}
	}
}
