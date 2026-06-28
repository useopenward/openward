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
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	APIKey    string             `json:"api_key"`
	Enabled   bool               `json:"enabled"`
	Upstream  string             `json:"upstream"`
	Algorithm RateLimitAlgorithm `json:"algorithm"`

	// Fixed window
	FWLimit  *int           `json:"fw_limit"`
	FWWindow *time.Duration `json:"fw_window"`

	// Sliding window
	SWLimit  *int           `json:"sw_limit"`
	SWWindow *time.Duration `json:"sw_window"`

	// Token bucket
	TBCapacity   *int     `json:"tb_capacity"`
	TBRefillRate *float64 `json:"tb_refill_rate"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
