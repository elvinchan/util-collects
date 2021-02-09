package ttl

import (
	"fmt"
	"testing"
	"time"
)

func TestCounter(t *testing.T) {
	ttl := NewCounter(time.Second * 5)
	ttl.Incr()
	time.Sleep(time.Second * 2)
	ttl.Incr()
	ticker := time.NewTicker(time.Second * 1)
	for {
		<-ticker.C
		fmt.Println("current len:", ttl.Len())
		if ttl.Len() == 0 {
			break
		}
	}
}
