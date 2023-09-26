package pq_test

import (
	"testing"

	"github.com/elvinchan/util-collects/as"
	"github.com/elvinchan/util-collects/container/counter/pq"
)

func TestPriorityQueue(t *testing.T) {
	q := pq.New(2, 3)
	type Value struct {
		Name string
	}
	itemA := q.Add("a", &Value{Name: "a"}, 2)
	itemB := q.Add("b", &Value{Name: "b"}, 3)
	q.Incr(itemA, 2)

	items := q.List()
	as.Equal(t, 2, len(items))
	as.Equal(t, items[0], itemA)
	as.Equal(t, items[0].Key(), "a")
	as.Equal(t, items[0].Priority(), int64(4))
	as.Equal(t, items[1], itemB)
	as.Equal(t, items[1].Key(), "b")
	as.Equal(t, items[1].Priority(), int64(3))

	_ = q.Add("c", &Value{Name: "c"}, 1)
	items = q.List()
	as.Equal(t, 3, len(items))
	as.Equal(t, items[0], itemA)
	as.Equal(t, items[0].Key(), "a")
	as.Equal(t, items[0].Priority(), int64(4))
	as.Equal(t, items[1], itemB)
	as.Equal(t, items[1].Key(), "b")
	as.Equal(t, items[1].Priority(), int64(3))

	itemD := q.Add("d", &Value{Name: "d"}, 6)
	items = q.List()
	as.Equal(t, 2, len(items))
	as.Equal(t, items[0], itemD)
	as.Equal(t, items[0].Key(), "d")
	as.Equal(t, items[0].Priority(), int64(6))
	as.Equal(t, items[1], itemA)
	as.Equal(t, items[1].Key(), "a")
	as.Equal(t, items[1].Priority(), int64(4))

	itemE := q.Add("e", &Value{Name: "e"}, 5)
	items = q.List()
	as.Equal(t, 3, len(items))
	as.Equal(t, items[0], itemD)
	as.Equal(t, items[0].Key(), "d")
	as.Equal(t, items[0].Priority(), int64(6))
	as.Equal(t, items[1], itemE)
	as.Equal(t, items[1].Key(), "e")
	as.Equal(t, items[1].Priority(), int64(5))
}
