package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type RateLimit struct {
	last     time.Time
	interval time.Duration
	maxCalls int
	count    int
	mu       sync.Mutex
}

func NewRateLimit(interval time.Duration, maxCalls int) *RateLimit {
	return &RateLimit{
		last:     time.Now(),
		interval: interval,
		maxCalls: maxCalls,
		count:    0,
	}
}

func (r *RateLimit) check(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// stop if max calls reached
	if r.maxCalls > 0 && r.count >= r.maxCalls {
		return fmt.Errorf("rate limit reached: max %d calls", r.maxCalls)
	}

	timePassed := time.Since(r.last)
	if timePassed < r.interval {
		wait := r.interval - timePassed
		r.mu.Unlock() // unlock while waiting
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return ctx.Err()
		}
		r.mu.Lock() // re-lock
	}

	r.last = time.Now()
	r.count++
	return nil
}
