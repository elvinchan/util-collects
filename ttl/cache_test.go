package ttl

import (
	"runtime"
	"testing"
	"time"

	"github.com/elvinchan/util-collects/testkit"
)

func TestSetGet(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		c := NewCache()
		testkit.Assert(t, c.items != nil)
		testkit.Assert(t, c.shutdown != nil)

		c.Set(1, "hello")
		testkit.Assert(t, c.Len() == 1)
		v, ok := c.Get(1)
		testkit.Assert(t, ok)
		testkit.Assert(t, v.(string) == "hello")
	})

	t.Run("TTL", func(t *testing.T) {
		c := NewCache(CacheWithTTL(time.Millisecond))
		testkit.Assert(t, c.ttl == time.Second)
		testkit.Assert(t, c.ttlList != nil)

		initNum := runtime.NumGoroutine()
		c.Set(1, "hello")
		testkit.Assert(t, c.Len() == 1)
		v, ok := c.Get(1)
		testkit.Assert(t, ok)
		testkit.Assert(t, v.(string) == "hello")
		testkit.Assert(t, runtime.NumGoroutine() == initNum+1)

		time.Sleep(time.Millisecond * 1100)
		testkit.Assert(t, c.Len() == 0)
		_, ok = c.Get(1)
		testkit.Assert(t, !ok)
		testkit.Assert(t, runtime.NumGoroutine() == initNum)
	})

	t.Run("LRU", func(t *testing.T) {
		c := NewCache(CacheWithLRU(2))
		testkit.Assert(t, c.cap == 2)
		testkit.Assert(t, c.lruList != nil)

		c.Set(1, "hello")
		c.Set("a", "world")
		testkit.Assert(t, c.Len() == 2)
		_, ok := c.Get(1)
		testkit.Assert(t, ok)
		_, ok = c.Get("a")
		testkit.Assert(t, ok)
		_, ok = c.Get(1)
		testkit.Assert(t, ok)

		c.Set(2, "cache")
		testkit.Assert(t, c.Len() == 2)
		_, ok = c.Get(2)
		testkit.Assert(t, ok)
		_, ok = c.Get(1)
		testkit.Assert(t, ok)
		_, ok = c.Get("a")
		testkit.Assert(t, !ok)
	})

	t.Run("TTL+LRU", func(t *testing.T) {
		c := NewCache(CacheWithTTL(time.Millisecond), CacheWithLRU(2))
		testkit.Assert(t, c.ttl == time.Second)
		testkit.Assert(t, c.ttlList != nil)
		testkit.Assert(t, c.cap == 2)
		testkit.Assert(t, c.lruList != nil)

		initNum := runtime.NumGoroutine()
		c.Set(1, "hello")
		c.Set("a", "world")
		testkit.Assert(t, c.Len() == 2)
		_, ok := c.Get(1)
		testkit.Assert(t, ok)
		_, ok = c.Get("a")
		testkit.Assert(t, ok)
		_, ok = c.Get(1)
		testkit.Assert(t, ok)

		c.Set(2, "cache")
		testkit.Assert(t, c.Len() == 2)
		_, ok = c.Get(2)
		testkit.Assert(t, ok)
		_, ok = c.Get(1)
		testkit.Assert(t, ok)
		_, ok = c.Get("a")
		testkit.Assert(t, !ok)
		testkit.Assert(t, runtime.NumGoroutine() == initNum+1)

		time.Sleep(time.Millisecond * 1100)
		testkit.Assert(t, c.Len() == 0)
		_, ok = c.Get(1)
		testkit.Assert(t, !ok)
		testkit.Assert(t, runtime.NumGoroutine() == initNum)
	})
}

func TestClose(t *testing.T) {
	c := NewCache(CacheWithTTL(time.Hour), CacheWithLRU(2))

	initNum := runtime.NumGoroutine()
	c.Set(1, "hello")
	testkit.Assert(t, c.Len() == 1)
	v, ok := c.Get(1)
	testkit.Assert(t, ok)
	testkit.Assert(t, v.(string) == "hello")
	testkit.Assert(t, runtime.NumGoroutine() == initNum+1)

	c.Close()
	time.Sleep(time.Millisecond)
	testkit.Assert(t, c.items == nil)
	testkit.Assert(t, runtime.NumGoroutine() == initNum)
}

func TestReset(t *testing.T) {
	c := NewCache(CacheWithTTL(time.Millisecond), CacheWithLRU(2))

	c.Set(1, "hello")
	testkit.Assert(t, c.Len() == 1)
	v, ok := c.Get(1)
	testkit.Assert(t, ok)
	testkit.Assert(t, v.(string) == "hello")

	c.Reset()
	testkit.Assert(t, c.Len() == 0)
	testkit.Assert(t, c.lruList.Len() == 0)
	testkit.Assert(t, c.ttlList.Len() == 0)
}
