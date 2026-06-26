package core

import (
	"fmt"
	"time"
)

type RateLimitAlgorithm string

const (
	AlgoFixedWindow   RateLimitAlgorithm = "fixed_window"
	AlgoSlidingWindow RateLimitAlgorithm = "sliding_window"
	AlgoTokenBucket   RateLimitAlgorithm = "token_bucket"
)

type Project struct {
	ID        string
	Name      string
	APIKey    string
	Enabled   bool
	Upstream  string
	Algorithm RateLimitAlgorithm

	// Fixed window
	// limit = max requests, window = duration of the window
	FWLimit  *int
	FWWindow *time.Duration

	// Sliding window
	// limit = max requests, window = duration of the window
	SWLimit  *int
	SWWindow *time.Duration

	// Token bucket:
	// TBCapacity     = max burst size (bucket size)
	// TBRefillRate   = tokens added per second (long-term average rate)
	TBCapacity   *int
	TBRefillRate *float64 // tokens/sec, e.g. 100.0 = 100 req/s sustained

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p *Project) FixedWindowConfig() (*FixedWindowConfig, error) {
	if p.Algorithm != AlgoFixedWindow {
		return nil, fmt.Errorf("project %s: algorithm is %s, not fixed_window", p.ID, p.Algorithm)
	}
	if p.FWLimit == nil || p.FWWindow == nil {
		return nil, fmt.Errorf("project %s: incomplete fixed_window config", p.ID)
	}
	return &FixedWindowConfig{Limit: *p.FWLimit, Window: *p.FWWindow}, nil
}

func (p *Project) SlidingWindowConfig() (*SlidingWindowConfig, error) {
	if p.Algorithm != AlgoSlidingWindow {
		return nil, fmt.Errorf("project %s: algorithm is %s, not sliding_window", p.ID, p.Algorithm)
	}
	if p.SWLimit == nil || p.SWWindow == nil {
		return nil, fmt.Errorf("project %s: incomplete sliding_window config", p.ID)
	}
	return &SlidingWindowConfig{Limit: *p.SWLimit, Window: *p.SWWindow}, nil
}

func (p *Project) TokenBucketConfig() (*TokenBucketConfig, error) {
	if p.Algorithm != AlgoTokenBucket {
		return nil, fmt.Errorf("project %s: algorithm is %s, not token_bucket", p.ID, p.Algorithm)
	}
	if p.TBCapacity == nil || p.TBRefillRate == nil {
		return nil, fmt.Errorf("project %s: incomplete token_bucket config", p.ID)
	}
	return &TokenBucketConfig{Capacity: *p.TBCapacity, RefillRate: *p.TBRefillRate}, nil
}

// Typed config structs — used by the limiter logic, not stored directly
type FixedWindowConfig struct {
	Limit  int
	Window time.Duration
}

type SlidingWindowConfig struct {
	Limit  int
	Window time.Duration
}

type TokenBucketConfig struct {
	Capacity   int
	RefillRate float64
}
