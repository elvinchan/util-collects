package ttl

import (
	"time"
)

func init() {
	defaultMinGap = time.Millisecond * 10
}
