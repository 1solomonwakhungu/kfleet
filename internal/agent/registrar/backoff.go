package registrar

import (
	"math/rand/v2"
	"time"
)

// Backoff produces capped exponential retry delays with jitter.
type Backoff struct {
	Base        time.Duration
	Max         time.Duration
	factor      float64
	current     time.Duration
	randFloat64 func() float64
}

// NewBackoff returns a backoff starting at one second and capped at one minute.
func NewBackoff() *Backoff {
	return &Backoff{Base: time.Second, Max: time.Minute, factor: 2, randFloat64: rand.Float64}
}

// Next returns the next delay with +/-20 percent jitter.
func (b *Backoff) Next() time.Duration {
	if b.current <= 0 {
		b.current = b.Base
	} else {
		next := time.Duration(float64(b.current) * b.factor)
		if next > b.Max || next < b.current {
			next = b.Max
		}
		b.current = next
	}
	randFloat64 := b.randFloat64
	if randFloat64 == nil {
		randFloat64 = rand.Float64
	}
	jitter := 0.8 + randFloat64()*0.4
	return time.Duration(float64(b.current) * jitter)
}

// Reset resets the next delay to the base duration.
func (b *Backoff) Reset() {
	b.current = 0
}
