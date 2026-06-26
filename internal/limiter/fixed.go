// internal/limiter/fixed.go
package limiter

import (
	"time"

	"github.com/useopenward/openward/internal/core"
	"github.com/useopenward/openward/internal/db"
)

type fixedWindow struct {
	cfg *core.FixedWindowConfig
}

// Allow counts requests in the current fixed window.
// The window start is floor(now / window) * window — aligns to wall clock.
func (f *fixedWindow) Allow(reader db.Reader, _ db.Writer, projectID string) (bool, error) {
	now := time.Now().UnixNano()
	windowNs := int64(f.cfg.Window)
	windowStart := (now / windowNs) * windowNs // floor to window boundary

	var count int
	err := reader.QueryRow(`
		SELECT COUNT(*) FROM request_logs
		WHERE project_id   = ?
		  AND allowed      = 1
		  AND requested_at >= ?
	`, projectID, windowStart/1e9).Scan(&count) // convert to unix seconds
	if err != nil {
		return false, err
	}

	return count < f.cfg.Limit, nil
}
