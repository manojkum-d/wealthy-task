package main

import (
	"context"
	"log"
	"sync"
	"time"
)

type RateLimit struct {
	mu         sync.Mutex
	interval   time.Duration
	maxCalls   int
	tokens     int
	lastRefill time.Time
}

func NewRateLimit(maxCalls int, interval time.Duration) *RateLimit {
	return &RateLimit{
		interval:   interval,
		maxCalls:   maxCalls,
		tokens:     maxCalls,
		lastRefill: time.Now(),
	}
}

func (r *RateLimit) check(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for {
		now := time.Now()
		elapsed := now.Sub(r.lastRefill)
		if elapsed >= r.interval {
			r.tokens = r.maxCalls
			r.lastRefill = now
		}

		if r.tokens > 0 {
			r.tokens--
			return nil
		}

		wait := r.interval - elapsed
		r.mu.Unlock()

		// 👇 Print waiting log before sleeping
		log.Printf("Waiting %v for rate limit reset...\n", wait)

		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return ctx.Err()
		}
		r.mu.Lock()
	}
}

