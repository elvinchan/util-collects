package backoff

import (
	"math"
	"math/rand"
	"sync/atomic"
	"time"
)

// JitterFunc defines a function that calculates a time.Duration based on
// the given duration.
type JitterFunc func(duration time.Duration) time.Duration

// Full creates a JitterFunc that transforms a duration into a result
// duration in [0, n) randomly, where n is the given duration.
//
// The given generator is what is used to determine the random jitterFunc.
// If a nil generator is passed, a default one will be provided.
//
// Inspired by https://www.awsarchitectureblog.com/2015/03/backoff.html
func Full(generator *rand.Rand) JitterFunc {
	random := fallbackNewRandom(generator)

	return func(duration time.Duration) time.Duration {
		return time.Duration(random.Int63n(int64(duration)))
	}
}

// Equal creates a JitterFunc that transforms a duration into a result
// duration in [n/2, n) randomly, where n is the given duration.
//
// The given generator is what is used to determine the random jitterFunc.
// If a nil generator is passed, a default one will be provided.
//
// Inspired by https://www.awsarchitectureblog.com/2015/03/backoff.html
func Equal(generator *rand.Rand) JitterFunc {
	random := fallbackNewRandom(generator)

	return func(duration time.Duration) time.Duration {
		return (duration / 2) + time.Duration(random.Int63n(int64(duration))/2)
	}
}

// Deviation creates a JitterFunc that transforms a duration into a result
// duration that deviates from the input randomly by a given factor.
//
// The given generator is what is used to determine the random jitterFunc.
// If a nil generator is passed, a default one will be provided.
//
// Inspired by https://developers.google.com/api-client-library/java/google-http-java-client/backoff
func Deviation(generator *rand.Rand, factor float64) JitterFunc {
	random := fallbackNewRandom(generator)

	return func(duration time.Duration) time.Duration {
		min := int64(math.Floor(float64(duration) * (1 - factor)))
		max := int64(math.Ceil(float64(duration) * (1 + factor)))

		return time.Duration(random.Int63n(max-min) + min)
	}
}

// NormalDistribution creates a JitterFunc that transforms a duration into a
// result duration based on a normal distribution of the input and the given
// standard deviation.
//
// The given generator is what is used to determine the random jitterFunc.
// If a nil generator is passed, a default one will be provided.
func NormalDistribution(generator *rand.Rand, standardDeviation float64) JitterFunc {
	random := fallbackNewRandom(generator)

	return func(duration time.Duration) time.Duration {
		return time.Duration(random.NormFloat64()*standardDeviation + float64(duration))
	}
}

var seedCounter int64

// fallbackNewRandom returns the passed in random instance if it's not nil,
// and otherwise returns a new random instance seeded with the current time.
func fallbackNewRandom(random *rand.Rand) *rand.Rand {
	// Return the passed in value if it's already not null
	if random != nil {
		return random
	}

	newSeed := atomic.AddInt64(&seedCounter, 1)
	return rand.New(rand.NewSource(time.Now().UnixNano() + newSeed))
}
