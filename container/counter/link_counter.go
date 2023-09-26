package counter

import (
	"sync"

	"github.com/elvinchan/util-collects/container/counter/pq"
)

type LinkCounter struct {
	linkMapper linkMap
	pq         pq.PriorityQueue
	mu         sync.Mutex
}

type linkMap struct {
	Mappings map[string]int
	List     []string // for get linkId by index
}

// since we use uint64 as bucket type, it can hold 64 bit.
const bucketCap = 64

// bit map of keys and linkIds. key is bucket index, value is bucket.
// 64 LinkId in one bucket, assume there's 2048 bucket，which can hold 131072
// LinkId, and occupies 16KB.
type linkBuckets map[uint16]uint64

// NewLinkCounter create a link counter which holds the cap count of keys with
// the max hits. When items count in queue exceeded cap, pop items to just keep
// retention count of items. cap <= 0 means no cap limit for counter.
func NewLinkCounter(retention int, cap int) *LinkCounter {
	return &LinkCounter{
		linkMapper: linkMap{
			Mappings: make(map[string]int),
		},
		pq: pq.New(retention, cap),
	}
}

// Add add or update a key with hits and linkId.
func (lc *LinkCounter) Add(key string, hits int64, linkId string) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	idx, ok := lc.linkMapper.Mappings[linkId]
	if !ok {
		lc.linkMapper.Mappings[linkId] = len(lc.linkMapper.List)
		idx = len(lc.linkMapper.List)
		lc.linkMapper.List = append(lc.linkMapper.List, linkId)
	}
	item, ok := lc.pq.Get(key)
	if !ok {
		mc := make(linkBuckets)
		hitBucket(&mc, idx)
		lc.pq.Add(key, &mc, hits)
	} else {
		mc := item.Value().(*linkBuckets)
		hitBucket(mc, idx)
		lc.pq.Incr(item, hits)
	}
}

func hitBucket(lb *linkBuckets, mappingId int) {
	(*lb)[uint16(mappingId/bucketCap)] |= 1 << (mappingId % bucketCap)
}

// Range provide a iteration function which ranges all key with hits and linkIds.
func (lc *LinkCounter) Range(f func(key string, hits int64, linkIds []string) bool) {
	lc.mu.Lock()
	keys := lc.pq.Keys()
	lc.mu.Unlock()

	for _, key := range keys {
		ok, hits, linkIds := lc.Get(key)
		if !ok {
			continue
		}
		if !f(key, hits, linkIds) {
			break
		}
	}
}

type KeyCount struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

// Get list of key with count.
func (lc *LinkCounter) CountList() []KeyCount {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	ts := lc.pq.List()

	v := make([]KeyCount, len(ts))
	for i := range ts {
		v[i] = KeyCount{
			Key:   ts[i].Key(),
			Count: ts[i].Priority(),
		}
	}
	return v
}

// Get check key exist and retrieves hits and linkIds of the key.
func (lc *LinkCounter) Get(key string) (bool, int64, []string) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	t, ok := lc.pq.Get(key)
	if !ok {
		return false, 0, nil
	}
	var linkIds []string
	lb := t.Value().(*linkBuckets)
	for bucketKey, bucketValue := range *lb {
		for i := 0; i < bucketCap; i++ {
			if bucketValue&(1<<i) > 0 {
				// hit
				id := int(bucketKey)*bucketCap + i
				if int(id) >= len(lc.linkMapper.List) {
					continue
				}
				linkIds = append(linkIds, lc.linkMapper.List[id])
			}
		}
	}
	return true, t.Priority(), linkIds
}
