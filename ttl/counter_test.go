package ttl

import (
	"fmt"
	"testing"
	"time"

	"github.com/elvinchan/util-collects/testkit"
)

func TestCounter(t *testing.T) {
	c := NewCounter(time.Second * 5)
	c.Incr()
	time.Sleep(time.Second * 2)
	c.Incr()
	ticker := time.NewTicker(time.Second * 1)
	results := []int{2, 2, 2, 1, 0}
	for i := 0; i < len(results); i++ {
		<-ticker.C
		fmt.Println("-----", c.Len())
		testkit.Assert(t, c.Len() == results[i])
	}
}
