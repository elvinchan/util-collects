package backoff

import (
	"math"
	"math/rand"
	"testing"
	"time"
)

func TestFullJitter(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	jitter := Full(r)

	duration := 100 * time.Millisecond
	for i := 0; i < 100; i++ {
		result := jitter(duration)
		if result < 0 || result >= duration {
			t.Errorf("FullJitter result %v out of [0, %v) range", result, duration)
		}
	}
}

func TestFullJitterWithNil(t *testing.T) {
	jitter := Full(nil)
	duration := 100 * time.Millisecond

	for i := 0; i < 100; i++ {
		result := jitter(duration)
		if result < 0 || result >= duration {
			t.Errorf("FullJitter(nil) result %v out of [0, %v) range", result, duration)
		}
	}
}

func TestEqualJitter(t *testing.T) {
	r := rand.New(rand.NewSource(2))
	jitter := Equal(r)

	duration := 100 * time.Millisecond
	half := duration / 2
	for i := 0; i < 100; i++ {
		result := jitter(duration)
		if result < half || result >= duration {
			t.Errorf("EqualJitter result %v out of [%v, %v) range", result, half, duration)
		}
	}
}

func TestEqualJitterWithNil(t *testing.T) {
	jitter := Equal(nil)
	duration := 100 * time.Millisecond
	half := duration / 2

	for i := 0; i < 100; i++ {
		result := jitter(duration)
		if result < half || result >= duration {
			t.Errorf("EqualJitter(nil) result %v out of [%v, %v) range", result, half, duration)
		}
	}
}

func TestDeviationJitter(t *testing.T) {
	r := rand.New(rand.NewSource(3))
	jitter := Deviation(r, 0.2)

	duration := 100 * time.Millisecond
	min := time.Duration(math.Floor(0.8 * float64(duration)))
	max := time.Duration(math.Ceil(1.2 * float64(duration)))

	for i := 0; i < 100; i++ {
		result := jitter(duration)
		if result < min || result >= max {
			t.Errorf("DeviationJitter result %v out of [%v, %v) range", result, min, max)
		}
	}
}

func TestDeviationJitterWithNil(t *testing.T) {
	jitter := Deviation(nil, 0.3)
	duration := 100 * time.Millisecond
	min := time.Duration(math.Floor(0.7 * float64(duration)))
	max := time.Duration(math.Ceil(1.3 * float64(duration)))

	for i := 0; i < 100; i++ {
		result := jitter(duration)
		if result < min || result >= max {
			t.Errorf("DeviationJitter(nil) result %v out of [%v, %v) range", result, min, max)
		}
	}
}

func TestNormalDistributionJitter(t *testing.T) {
	r := rand.New(rand.NewSource(4))
	jitter := NormalDistribution(r, 10.0)

	duration := 100 * time.Millisecond
	for i := 0; i < 100; i++ {
		result := jitter(duration)
		if result < 0 {
			t.Errorf("NormalDistributionJitter got negative duration: %v", result)
		}
	}
}

func TestNormalDistributionJitterWithNil(t *testing.T) {
	jitter := NormalDistribution(nil, 5.0)
	duration := 100 * time.Millisecond

	for i := 0; i < 100; i++ {
		result := jitter(duration)
		if result < 0 {
			t.Errorf("NormalDistributionJitter(nil) got negative duration: %v", result)
		}
	}
}
