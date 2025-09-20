package rate

import (
	"sync"
	"time"
)

// Limiter is a per-key token bucket limiter with fixed rps and burst.
type Limiter struct {
	mu     sync.Mutex
	rps    float64
	burst  float64
	bucket map[string]*tb
}

type tb struct {
	tokens float64
	last   time.Time
}

// NewLimiter creates a limiter with rps tokens per second and the given burst.
func NewLimiter(rps, burst int) *Limiter {
	if rps <= 0 {
		rps = 10
	}
	if burst <= 0 {
		burst = rps
	}
	return &Limiter{
		rps:    float64(rps),
		burst:  float64(burst),
		bucket: make(map[string]*tb),
	}
}

// Allow consumes one token for key if available and returns true.
// Otherwise returns false.
func (l *Limiter) Allow(key string) bool {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.bucket[key]
	if !ok {
		l.bucket[key] = &tb{tokens: l.burst - 1, last: now}
		return true
	}

	// Refill based on elapsed time
	elapsed := now.Sub(b.last).Seconds()
	if elapsed > 0 {
		b.tokens = min(l.burst, b.tokens+elapsed*l.rps)
		b.last = now
	}

	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		return true
	}
	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
