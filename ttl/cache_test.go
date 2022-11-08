package ttl

import (
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/elvinchan/util-collects/as"
)

func TestSetGet(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		c := NewCache()
		as.NotEqual(t, c.items, nil)
		as.NotEqual(t, c.shutdown, nil)

		c.Set(1, "hello")
		as.Equal(t, c.Len(), 1)
		v, ok := c.Get(1)
		as.True(t, ok)
		as.Equal(t, v.(string), "hello")
	})

	t.Run("TTL", func(t *testing.T) {
		c := NewCache(CacheWithTTL(time.Millisecond))
		as.Equal(t, c.ttl, time.Millisecond*10)
		as.NotEqual(t, c.ttlList, nil)

		c.Set(1, "hello")
		as.Equal(t, c.Len(), 1)
		v, ok := c.Get(1)
		as.True(t, ok)
		as.Equal(t, v.(string), "hello")
		as.Equal(t, atomic.LoadUint32(&c.cleaning), uint32(1))

		time.Sleep(time.Millisecond * 20)
		as.Equal(t, c.Len(), 0)
		_, ok = c.Get(1)
		as.False(t, ok)
		as.Equal(t, atomic.LoadUint32(&c.cleaning), uint32(0))
	})

	t.Run("LRU", func(t *testing.T) {
		c := NewCache(CacheWithLRU(2))
		as.Equal(t, c.cap, 2)
		as.NotEqual(t, c.lruList, nil)

		c.Set(1, "hello")
		c.Set("a", "world")
		as.Equal(t, c.Len(), 2)
		_, ok := c.Get(1)
		as.True(t, ok)
		_, ok = c.Get("a")
		as.True(t, ok)
		_, ok = c.Get(1)
		as.True(t, ok)

		c.Set(2, "cache")
		as.Equal(t, c.Len(), 2)
		_, ok = c.Get(2)
		as.True(t, ok)
		_, ok = c.Get(1)
		as.True(t, ok)
		_, ok = c.Get("a")
		as.False(t, ok)
	})

	t.Run("TTL+LRU", func(t *testing.T) {
		c := NewCache(CacheWithTTL(time.Millisecond), CacheWithLRU(2))
		as.Equal(t, c.ttl, time.Millisecond*10)
		as.NotEqual(t, c.ttlList, nil)
		as.Equal(t, c.cap, 2)
		as.NotEqual(t, c.lruList, nil)

		c.Set(1, "hello")
		c.Set("a", "world")
		as.Equal(t, c.Len(), 2)
		_, ok := c.Get(1)
		as.True(t, ok)
		_, ok = c.Get("a")
		as.True(t, ok)
		_, ok = c.Get(1)
		as.True(t, ok)

		c.Set(2, "cache")
		as.Equal(t, c.Len(), 2)
		_, ok = c.Get(2)
		as.True(t, ok)
		_, ok = c.Get(1)
		as.True(t, ok)
		_, ok = c.Get("a")
		as.False(t, ok)
		as.Equal(t, atomic.LoadUint32(&c.cleaning), uint32(1))

		time.Sleep(time.Millisecond * 20)
		as.Equal(t, c.Len(), 0)
		_, ok = c.Get(1)
		as.False(t, ok)
		as.Equal(t, atomic.LoadUint32(&c.cleaning), uint32(0))
	})
}

func TestClose(t *testing.T) {
	c := NewCache(CacheWithTTL(time.Hour), CacheWithLRU(2))

	initNum := runtime.NumGoroutine()
	c.Set(1, "hello")
	as.Equal(t, c.Len(), 1)
	v, ok := c.Get(1)
	as.True(t, ok)
	as.Equal(t, v.(string), "hello")
	as.Equal(t, runtime.NumGoroutine(), initNum+1)

	c.Close()
	time.Sleep(time.Millisecond)
	as.Equal(t, len(c.items), 0)
	as.Equal(t, runtime.NumGoroutine(), initNum)
}
