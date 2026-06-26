// internal/limiter/sliding.go
package limiter

import (
	"time"

	"github.com/useopenward/openward/internal/core"
	"github.com/useopenward/openward/internal/db"
)

type slidingWindow struct {
	cfg *core.SlidingWindowConfig
}

// Allow uses the approximated sliding window algorithm:
//
//	approximation = (prevWindowCount * prevWindowWeight) + currentWindowCount
//
// This avoids storing per-request timestamps while still smoothing bursts.
func (s *slidingWindow) Allow(reader db.Reader, _ db.Writer, projectID string) (bool, error) {
	now := time.Now()
	windowNs := int64(s.cfg.Window)
	nowNs := now.UnixNano()

	currentWindowStart := (nowNs / windowNs) * windowNs
	prevWindowStart := currentWindowStart - windowNs

	// how far into the current window are we? (0.0 → 1.0)
	elapsed := float64(nowNs-currentWindowStart) / float64(windowNs)
	prevWeight := 1.0 - elapsed

	var prevCount, currentCount int

	err := reader.QueryRow(`
		SELECT COUNT(*) FROM request_logs
		WHERE project_id   = ?
		  AND allowed      = 1
		  AND requested_at >= ? AND requested_at < ?
	`, projectID, prevWindowStart/1e9, currentWindowStart/1e9).Scan(&prevCount)
	if err != nil {
		return false, err
	}

	err = reader.QueryRow(`
		SELECT COUNT(*) FROM request_logs
		WHERE project_id   = ?
		  AND allowed      = 1
		  AND requested_at >= ?
	`, projectID, currentWindowStart/1e9).Scan(&currentCount)
	if err != nil {
		return false, err
	}

	approximation := (float64(prevCount) * prevWeight) + float64(currentCount)
	return approximation < float64(s.cfg.Limit), nil
}
