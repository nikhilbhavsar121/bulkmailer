package limiter

import (
	"sync"
	"time"
)

const DefaultRateLimit = 1000

type SlidingRateLimiter struct {
	mu      sync.Mutex
	limits  map[string]int
	history map[string][]time.Time
	window  time.Duration
}

func NewSlidingRateLimiter(window time.Duration) *SlidingRateLimiter {
	return &SlidingRateLimiter{
		limits:  make(map[string]int),
		history: make(map[string][]time.Time),
		window:  window,
	}
}

func (r *SlidingRateLimiter) Allow(tpn string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	limit := r.limits[tpn]
	if limit == 0 {
		limit = DefaultRateLimit
	}

	now := time.Now()
	cutoff := now.Add(-r.window)
	valid := r.history[tpn][:0]
	for _, t := range r.history[tpn] {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	r.history[tpn] = valid

	if len(valid) >= limit {
		return false
	}

	r.history[tpn] = append(r.history[tpn], now)
	return true
}

func (r *SlidingRateLimiter) SetLimit(tpn string, limit int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.limits[tpn] = limit
}
