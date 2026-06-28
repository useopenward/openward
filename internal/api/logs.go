package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"
)

type requestLog struct {
	ID          int64     `json:"id"`
	ProjectID   string    `json:"project_id"`
	RequestedAt time.Time `json:"requested_at"`
	Allowed     bool      `json:"allowed"`
	StatusCode  *int      `json:"status_code"`
}

func (s *Server) handleListLogs(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= 500 {
			limit = v
		}
	}

	offsetStr := r.URL.Query().Get("offset")
	offset := 0
	if offsetStr != "" {
		if v, err := strconv.Atoi(offsetStr); err == nil && v >= 0 {
			offset = v
		}
	}

	rows, err := s.db.Reader.Query(`
		SELECT id, project_id, requested_at, allowed, status_code
		FROM request_logs
		WHERE project_id = ?
		ORDER BY requested_at DESC
		LIMIT ? OFFSET ?
	`, id, limit, offset)
	if err != nil {
		http.Error(w, "failed to query logs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logs []requestLog
	for rows.Next() {
		var l requestLog
		var requestedAt int64
		var statusCode sql.NullInt64
		var allowed int

		if err := rows.Scan(&l.ID, &l.ProjectID, &requestedAt, &allowed, &statusCode); err != nil {
			http.Error(w, "failed to scan logs", http.StatusInternalServerError)
			return
		}
		l.RequestedAt = time.Unix(requestedAt, 0)
		l.Allowed = allowed == 1
		if statusCode.Valid {
			v := int(statusCode.Int64)
			l.StatusCode = &v
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "failed to read logs", http.StatusInternalServerError)
		return
	}

	writeJSON(w, logs)
}
