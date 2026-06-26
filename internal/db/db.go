package db

import (
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

type Handles struct {
	Reader *sql.DB
	Writer *sql.DB
}

func (h *Handles) Close() error {
	var errs []error
	if h.Reader != nil {
		if err := h.Reader.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if h.Writer != nil {
		if err := h.Writer.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return joinErrors(errs...)
}

type Reader interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

type Writer interface {
	Exec(query string, args ...any) (sql.Result, error)
}

func Open(path string) (*Handles, error) {
	writer, err := sql.Open("sqlite", dsn(path))
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	writer.SetMaxOpenConns(1)
	writer.SetMaxIdleConns(1)

	if err := writer.Ping(); err != nil {
		_ = writer.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if err := migrate(writer); err != nil {
		_ = writer.Close()
		return nil, fmt.Errorf("migrate db: %w", err)
	}

	reader, err := sql.Open("sqlite", dsn(path))
	if err != nil {
		_ = writer.Close()
		return nil, fmt.Errorf("open reader db: %w", err)
	}

	reader.SetMaxOpenConns(4)
	reader.SetMaxIdleConns(4)

	if err := reader.Ping(); err != nil {
		_ = reader.Close()
		_ = writer.Close()
		return nil, fmt.Errorf("ping reader db: %w", err)
	}

	return &Handles{
		Reader: reader,
		Writer: writer,
	}, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(schema)
	return err
}

func dsn(path string) string {
	return path + "?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)"
}

func joinErrors(errs ...error) error {
	return errors.Join(errs...)
}

const schema = `
CREATE TABLE IF NOT EXISTS projects (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    api_key     TEXT NOT NULL UNIQUE,
    enabled     INTEGER NOT NULL DEFAULT 1,
    upstream    TEXT NOT NULL,
    algorithm   TEXT NOT NULL,

    -- fixed window
    fw_limit    INTEGER,
    fw_window   INTEGER,           -- nanoseconds (time.Duration)

    -- sliding window
    sw_limit    INTEGER,
    sw_window   INTEGER,           -- nanoseconds (time.Duration)

    -- token bucket
    tb_capacity    INTEGER,
    tb_refill_rate REAL,           -- tokens/sec

    created_at  INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at  INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE TABLE IF NOT EXISTS request_logs (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id   TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    requested_at INTEGER NOT NULL DEFAULT (unixepoch()),
    allowed      INTEGER NOT NULL,   -- 0 or 1
    status_code  INTEGER             -- null if blocked before upstream
);

CREATE TABLE IF NOT EXISTS token_bucket_state (
    project_id  TEXT PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
    last_time   INTEGER NOT NULL,   -- unix seconds of last request
    tokens      REAL    NOT NULL    -- current token count
);

CREATE INDEX IF NOT EXISTS idx_request_logs_project_time
    ON request_logs(project_id, requested_at DESC);
`
