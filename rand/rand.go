package rand

import (
	"math/rand"
	"time"
)

func RandInt64(min, max int64) int64 {
	rand.Seed(time.Now().UnixNano())
	if min >= max {
		return max
	}
	return rand.Int63n(max-min) + min
}

func RandInt64NoSeed(min, max int64) int64 {
	if min >= max {
		return max
	}
	return rand.Int63n(max-min) + min
}

func RandInt32(min, max int32) int32 {
	rand.Seed(time.Now().UnixNano())
	if min >= max {
		return max
	}
	return rand.Int31n(max-min) + min
}

func RandInt32NoSeed(min, max int32) int32 {
	if min >= max {
		return max
	}
	return rand.Int31n(max-min) + min
}

func RandInt(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	if min >= max {
		return max
	}
	return rand.Intn(max-min) + min
}

func RandIntNoSeed(min, max int) int {
	if min >= max {
		return max
	}
	return rand.Intn(max-min) + min
}
