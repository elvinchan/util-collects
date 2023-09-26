package pq

import (
	"container/heap"
	"sort"
)

// based on
// https://github.com/golang/go/blob/master/src/container/heap/example_pq_test.go

// An Item is something we manage in a priority queue.
type Item struct {
	key      string      // key for get item directly; immutable.
	value    interface{} // The value of the item; arbitrary.
	priority int64       // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

func (t *Item) Key() string {
	return t.key
}

func (t *Item) Value() interface{} {
	return t.value
}

func (t *Item) Priority() int64 {
	return t.priority
}

// A PriorityQueue implements heap.Interface and holds Items with key dictionary.
type priorityQueue struct {
	keyDict   map[string]*Item
	items     []*Item
	retention int
	cap       int
}

func (pq priorityQueue) Len() int { return len(pq.items) }

func (pq priorityQueue) Less(i, j int) bool {
	// We want Pop to give us the lowest, not highest, priority so we use less than here.
	return pq.items[i].priority < pq.items[j].priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].index = i
	pq.items[j].index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len((*pq).items)
	item := x.(*Item)
	item.index = n
	(*pq).items = append((*pq).items, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old.items)
	item := old.items[n-1]
	old.items[n-1] = nil // avoid memory leak
	item.index = -1      // for safety
	(*pq).items = old.items[0 : n-1]
	delete(pq.keyDict, item.key)
	return item
}

func (pq *priorityQueue) Add(key string, value interface{}, priority int64) *Item {
	if t, ok := pq.keyDict[key]; ok {
		return t
	}
	item := &Item{
		key:      key,
		value:    value,
		priority: priority,
	}
	pq.keyDict[key] = item
	heap.Push(pq, item)
	if pq.cap > 0 && pq.Len() > pq.cap {
		for pq.Len() > pq.retention {
			heap.Pop(pq)
		}
	}
	return item
}

// Incr increase the priority of an Item in the queue.
func (pq *priorityQueue) Incr(item *Item, priority int64) {
	item.priority += priority
	heap.Fix(pq, item.index)
}

// Get retrieves Item of key.
func (pq *priorityQueue) Get(key string) (*Item, bool) {
	item, ok := pq.keyDict[key]
	return item, ok
}

// Keys get all keys of the queue.
func (pq *priorityQueue) Keys() []string {
	keys := make([]string, len(pq.items))
	for i, t := range pq.items {
		keys[i] = t.key
	}
	return keys
}

// List get all Items of the queue.
func (pq *priorityQueue) List() []*Item {
	v := make([]*Item, pq.Len())
	copy(v, pq.items)
	sort.Slice(v, func(i, j int) bool {
		return v[i].priority > v[j].priority
	})
	return v
}

type PriorityQueue interface {
	Add(key string, value interface{}, priority int64) *Item
	Incr(item *Item, priority int64)
	Get(key string) (*Item, bool)
	Keys() []string
	List() []*Item
}

// New create a PriorityQueue with retention and cap, when items count in queue
// exceeded cap, pop items to just keep retention count of items.
// cap <= 0 means no cap limit for PriorityQueue.
func New(retention, cap int) PriorityQueue {
	q := priorityQueue{
		keyDict:   make(map[string]*Item),
		retention: retention,
		cap:       cap,
	}
	return &q
}
