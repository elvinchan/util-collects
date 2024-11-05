package ttl

import (
	"time"
)

const TestDefaultMinGap = time.Millisecond * 100
const Testjitter = time.Millisecond * 50

func init() {
	defaultMinGap = TestDefaultMinGap
}
