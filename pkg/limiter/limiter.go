package limiter

import (
	"sync"
	"time"
)

// RateLimiter defines interface for rate limiting
type RateLimiter interface {
	// Allow checks if the action is allowed for the key.
	// Returns true if allowed, false otherwise.
	Allow(key string) bool
}

// InMemoryLimiter implements a simple window-based rate limiter
type InMemoryLimiter struct {
	mu       sync.Mutex
	counters map[string]*window
	limit    int
	window   time.Duration
}

type window struct {
	count     int
	startTime time.Time
}

func NewInMemoryLimiter(limit int, windowDuration time.Duration) *InMemoryLimiter {
	l := &InMemoryLimiter{
		counters: make(map[string]*window),
		limit:    limit,
		window:   windowDuration,
	}
	// Start cleanup routine
	go l.cleanup()
	return l
}

func (l *InMemoryLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	w, exists := l.counters[key]

	if !exists || now.Sub(w.startTime) > l.window {
		// New window or expired window
		l.counters[key] = &window{
			count:     1,
			startTime: now,
		}
		return true
	}

	if w.count >= l.limit {
		return false
	}

	w.count++
	return true
}

func (l *InMemoryLimiter) cleanup() {
	ticker := time.NewTicker(l.window * 2)
	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for k, w := range l.counters {
			if now.Sub(w.startTime) > l.window {
				delete(l.counters, k)
			}
		}
		l.mu.Unlock()
	}
}
