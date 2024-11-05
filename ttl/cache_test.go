package ttl

import (
	"context"
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
		c := NewCache(CacheWithTTL(time.Millisecond, TTLRefreshModeNone))
		as.Equal(t, c.ttl, TestDefaultMinGap)
		as.NotEqual(t, c.ttlList, nil)

		c.Set(1, "hello")
		as.Equal(t, c.Len(), 1)
		v, ok := c.Get(1)
		as.True(t, ok)
		as.Equal(t, v.(string), "hello")
		as.Equal(t, atomic.LoadUint32(&c.cleaning), uint32(1))

		as.Eventually(t, func(ctx context.Context) bool {
			return c.Len() == 0
		}, TestDefaultMinGap+Testjitter, TestDefaultMinGap)
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
		c := NewCache(CacheWithTTL(time.Millisecond, TTLRefreshModeNone),
			CacheWithLRU(2))
		as.Equal(t, c.ttl, TestDefaultMinGap)
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

		as.Eventually(t, func(ctx context.Context) bool {
			return c.Len() == 0
		}, TestDefaultMinGap+Testjitter, TestDefaultMinGap)
		_, ok = c.Get(1)
		as.False(t, ok)
		as.Equal(t, atomic.LoadUint32(&c.cleaning), uint32(0))
	})
}

func TestClose(t *testing.T) {
	c := NewCache(CacheWithTTL(time.Hour, TTLRefreshModeNone), CacheWithLRU(2))

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

func TestTTL(t *testing.T) {
	t.Run("TTLGet", func(t *testing.T) {
		gap := TestDefaultMinGap * 2
		c := NewCache(CacheWithTTL(gap, TTLRefreshModeGet))

		c.Set(1, "hello")
		as.Equal(t, c.Len(), 1)
		expireAt := c.items[1].expireAt
		as.True(t, expireAt.After(time.Now()))
		as.True(t, expireAt.Before(time.Now().Add(gap)))
		as.False(t, expireAt.IsZero())

		// shift == 0

		time.Sleep(TestDefaultMinGap)
		v, ok := c.Get(1)
		as.True(t, ok)
		as.Equal(t, v.(string), "hello")
		as.Equal(t, atomic.LoadUint32(&c.cleaning), uint32(1))

		// shift == +100ms

		nextExpireAt := c.items[1].expireAt
		as.True(t, nextExpireAt.After(expireAt))
		as.True(t, nextExpireAt.After(time.Now()))
		as.True(t, nextExpireAt.Before(time.Now().Add(gap)))

		// shift == +100ms

		as.Never(t, func(ctx context.Context) bool {
			return c.Len() == 0
		}, gap, TestDefaultMinGap/2)

		// shift == +300ms

		as.Eventually(t, func(ctx context.Context) bool {
			return c.Len() == 0
		}, gap, TestDefaultMinGap/2)

		// shift <= +500ms

		_, ok = c.Get(1)
		as.False(t, ok)
		as.Equal(t, atomic.LoadUint32(&c.cleaning), uint32(0))
	})

	t.Run("TTLSet", func(t *testing.T) {
		gap := TestDefaultMinGap * 2
		c := NewCache(CacheWithTTL(gap, TTLRefreshModeSet))

		c.Set(1, "hello")
		as.Equal(t, c.Len(), 1)
		expireAt := c.items[1].expireAt
		as.True(t, expireAt.After(time.Now()))
		as.True(t, expireAt.Before(time.Now().Add(gap)))
		as.False(t, expireAt.IsZero())

		// shift == 0

		time.Sleep(TestDefaultMinGap)
		c.Set(1, "world")
		as.Equal(t, atomic.LoadUint32(&c.cleaning), uint32(1))

		// shift == +100ms

		nextExpireAt := c.items[1].expireAt
		as.True(t, nextExpireAt.After(expireAt))
		as.True(t, nextExpireAt.After(time.Now()))
		as.True(t, nextExpireAt.Before(time.Now().Add(gap)))

		// shift == +100ms

		as.Never(t, func(ctx context.Context) bool {
			return c.Len() == 0
		}, gap, TestDefaultMinGap/2)

		// shift == +300ms

		as.Eventually(t, func(ctx context.Context) bool {
			return c.Len() == 0
		}, gap, TestDefaultMinGap/2)

		// shift <= +500ms

		_, ok := c.Get(1)
		as.False(t, ok)
		as.Equal(t, atomic.LoadUint32(&c.cleaning), uint32(0))
	})

	t.Run("TTLGetSet", func(t *testing.T) {
		gap := TestDefaultMinGap * 2
		c := NewCache(CacheWithTTL(gap, TTLRefreshModeGet|TTLRefreshModeSet))

		c.Set(1, "hello")
		as.Equal(t, c.Len(), 1)
		expireAt := c.items[1].expireAt
		as.True(t, expireAt.After(time.Now()))
		as.True(t, expireAt.Before(time.Now().Add(gap)))
		as.False(t, expireAt.IsZero())

		// shift == 0

		time.Sleep(TestDefaultMinGap)
		v, ok := c.Get(1)
		as.True(t, ok)
		as.Equal(t, v.(string), "hello")
		as.Equal(t, atomic.LoadUint32(&c.cleaning), uint32(1))

		// shift == +100ms

		nextExpireAt := c.items[1].expireAt
		as.True(t, nextExpireAt.After(expireAt))
		as.True(t, nextExpireAt.After(time.Now()))
		as.True(t, nextExpireAt.Before(time.Now().Add(gap)))

		// shift == +100ms

		as.Never(t, func(ctx context.Context) bool {
			return c.Len() == 0
		}, gap, TestDefaultMinGap/2)

		// shift == +300ms

		as.Eventually(t, func(ctx context.Context) bool {
			return c.Len() == 0
		}, gap, TestDefaultMinGap/2)

		// shift <= +500ms

		_, ok = c.Get(1)
		as.False(t, ok)
		as.Equal(t, atomic.LoadUint32(&c.cleaning), uint32(0))

		// set again

		c.Set(1, "hello")
		as.Equal(t, c.Len(), 1)
		expireAt = c.items[1].expireAt
		as.True(t, expireAt.After(time.Now()))
		as.True(t, expireAt.Before(time.Now().Add(gap)))
		as.False(t, expireAt.IsZero())

		// shift == 0

		time.Sleep(TestDefaultMinGap)
		c.Set(1, "world")
		as.Equal(t, atomic.LoadUint32(&c.cleaning), uint32(1))

		// shift == +100ms

		nextExpireAt = c.items[1].expireAt
		as.True(t, nextExpireAt.After(expireAt))
		as.True(t, nextExpireAt.After(time.Now()))
		as.True(t, nextExpireAt.Before(time.Now().Add(gap)))

		// shift == +100ms

		as.Never(t, func(ctx context.Context) bool {
			return c.Len() == 0
		}, gap, TestDefaultMinGap/2)

		// shift == +300ms

		as.Eventually(t, func(ctx context.Context) bool {
			return c.Len() == 0
		}, gap, TestDefaultMinGap/2)

		// shift <= +500ms

		_, ok = c.Get(1)
		as.False(t, ok)
		as.Equal(t, atomic.LoadUint32(&c.cleaning), uint32(0))
	})

	t.Run("TTLNone", func(t *testing.T) {
		gap := TestDefaultMinGap * 2
		c := NewCache(CacheWithTTL(gap, TTLRefreshModeNone))

		c.Set(1, "hello")
		as.Equal(t, c.Len(), 1)
		expireAt := c.items[1].expireAt
		as.True(t, expireAt.After(time.Now()))
		as.True(t, expireAt.Before(time.Now().Add(gap)))
		as.False(t, expireAt.IsZero())

		// shift == 0

		time.Sleep(gap - Testjitter)
		v, ok := c.Get(1)
		as.True(t, ok)
		as.Equal(t, v.(string), "hello")

		// shift ~= +200ms

		nextExpireAt := c.items[1].expireAt
		as.Equal(t, nextExpireAt, expireAt)

		// shift ~= +200ms

		as.Eventually(t, func(ctx context.Context) bool {
			return c.Len() == 0
		}, gap, TestDefaultMinGap/2)

		// shift ~= +400ms

		_, ok = c.Get(1)
		as.False(t, ok)
		as.Equal(t, atomic.LoadUint32(&c.cleaning), uint32(0))

		// set again

		c.Set(1, "hello")
		as.Equal(t, c.Len(), 1)
		expireAt = c.items[1].expireAt
		as.True(t, expireAt.After(time.Now()))
		as.True(t, expireAt.Before(time.Now().Add(gap)))
		as.False(t, expireAt.IsZero())

		// shift == 0

		time.Sleep(gap - Testjitter)
		c.Set(1, "world")

		// shift ~= +200ms

		nextExpireAt = c.items[1].expireAt
		as.Equal(t, nextExpireAt, expireAt)

		// shift ~= +200ms

		as.Eventually(t, func(ctx context.Context) bool {
			return c.Len() == 0
		}, gap, TestDefaultMinGap/2)

		// shift ~= +400ms

		_, ok = c.Get(1)
		as.False(t, ok)
		as.Equal(t, atomic.LoadUint32(&c.cleaning), uint32(0))
	})
}
