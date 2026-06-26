// internal/limiter/limiter.go
package limiter

import (
	"fmt"

	"github.com/useopenward/openward/internal/core"
	"github.com/useopenward/openward/internal/db"
)

type Limiter interface {
	Allow(reader db.Reader, writer db.Writer, projectID string) (bool, error)
}

func New(p *core.Project) (Limiter, error) {
	switch p.Algorithm {
	case core.AlgoFixedWindow:
		cfg, err := p.FixedWindowConfig()
		if err != nil {
			return nil, err
		}
		return &fixedWindow{cfg: cfg}, nil

	case core.AlgoSlidingWindow:
		cfg, err := p.SlidingWindowConfig()
		if err != nil {
			return nil, err
		}
		return &slidingWindow{cfg: cfg}, nil

	case core.AlgoTokenBucket:
		cfg, err := p.TokenBucketConfig()
		if err != nil {
			return nil, err
		}
		return &tokenBucket{cfg: cfg}, nil

	default:
		return nil, fmt.Errorf("unknown algorithm: %s", p.Algorithm)
	}
}
