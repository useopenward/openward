// internal/limiter/token.go
package limiter

import (
	"database/sql"
	"time"

	"github.com/useopenward/openward/internal/core"
	"github.com/useopenward/openward/internal/db"
)

type tokenBucket struct {
	cfg *core.TokenBucketConfig
}

// Allow implements token bucket via virtual tokens —
// no background goroutine needed. We compute the current
// token count from elapsed time since the last request.
//
//	tokens = min(capacity, lastTokens + (elapsed * refillRate))
func (t *tokenBucket) Allow(reader db.Reader, writer db.Writer, projectID string) (bool, error) {
	now := time.Now().Unix()

	// fetch last state: most recent allowed request time + running token count
	// we store tokens in a separate table to avoid scanning request_logs
	var lastTime int64
	var lastTokens float64

	err := reader.QueryRow(`
		SELECT last_time, tokens FROM token_bucket_state
		WHERE project_id = ?
	`, projectID).Scan(&lastTime, &lastTokens)

	if err == sql.ErrNoRows {
		// first request — bucket starts full
		lastTime = now
		lastTokens = float64(t.cfg.Capacity)
	} else if err != nil {
		return false, err
	}

	// refill based on elapsed time
	elapsed := float64(now - lastTime)
	tokens := lastTokens + elapsed*t.cfg.RefillRate
	if tokens > float64(t.cfg.Capacity) {
		tokens = float64(t.cfg.Capacity)
	}

	if tokens < 1 {
		return false, nil // bucket empty
	}

	// consume one token and persist state
	tokens--
	_, err = writer.Exec(`
		INSERT INTO token_bucket_state (project_id, last_time, tokens)
		VALUES (?, ?, ?)
		ON CONFLICT(project_id) DO UPDATE SET
			last_time = excluded.last_time,
			tokens    = excluded.tokens
	`, projectID, now, tokens)
	if err != nil {
		return false, err
	}

	return true, nil
}
