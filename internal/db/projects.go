package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/useopenward/openward/internal/core"
)

var ErrNotFound = errors.New("not found")

// scanProject maps a db row onto a core.Project.
// Column order must match every SELECT that uses this.
func scanProject(row interface {
	Scan(...any) error
}) (*core.Project, error) {
	var p core.Project
	var (
		fwLimit, fwWindow    sql.NullInt64
		swLimit, swWindow    sql.NullInt64
		tbCapacity           sql.NullInt64
		tbRefillRate         sql.NullFloat64
		createdAt, updatedAt int64
	)

	err := row.Scan(
		&p.ID, &p.Name, &p.APIKey, &p.Enabled, &p.Upstream, &p.Algorithm,
		&fwLimit, &fwWindow,
		&swLimit, &swWindow,
		&tbCapacity, &tbRefillRate,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	if fwLimit.Valid {
		v := int(fwLimit.Int64)
		p.FWLimit = &v
	}
	if fwWindow.Valid {
		v := time.Duration(fwWindow.Int64)
		p.FWWindow = &v
	}
	if swLimit.Valid {
		v := int(swLimit.Int64)
		p.SWLimit = &v
	}
	if swWindow.Valid {
		v := time.Duration(swWindow.Int64)
		p.SWWindow = &v
	}
	if tbCapacity.Valid {
		v := int(tbCapacity.Int64)
		p.TBCapacity = &v
	}
	if tbRefillRate.Valid {
		p.TBRefillRate = &tbRefillRate.Float64
	}

	p.CreatedAt = time.Unix(createdAt, 0)
	p.UpdatedAt = time.Unix(updatedAt, 0)

	return &p, nil
}

const projectColumns = `
	id, name, api_key, enabled, upstream, algorithm,
	fw_limit, fw_window,
	sw_limit, sw_window,
	tb_capacity, tb_refill_rate,
	created_at, updated_at
`

func GetProject(db Reader, id string) (*core.Project, error) {
	row := db.QueryRow(
		`SELECT `+projectColumns+` FROM projects WHERE id = ?`, id,
	)
	p, err := scanProject(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return p, err
}

func GetProjectByAPIKey(db Reader, apiKey string) (*core.Project, error) {
	row := db.QueryRow(
		`SELECT `+projectColumns+` FROM projects WHERE api_key = ?`, apiKey,
	)
	p, err := scanProject(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return p, err
}

func ListProjects(db Reader) ([]*core.Project, error) {
	rows, err := db.Query(`SELECT ` + projectColumns + ` FROM projects ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := make([]*core.Project, 0)
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func CreateProject(db Writer, p *core.Project) error {
	_, err := db.Exec(`
		INSERT INTO projects (
			id, name, api_key, enabled, upstream, algorithm,
			fw_limit, fw_window,
			sw_limit, sw_window,
			tb_capacity, tb_refill_rate
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.APIKey, p.Enabled, p.Upstream, p.Algorithm,
		nullInt(p.FWLimit), nullDuration(p.FWWindow),
		nullInt(p.SWLimit), nullDuration(p.SWWindow),
		nullInt(p.TBCapacity), p.TBRefillRate,
	)
	if err != nil {
		return fmt.Errorf("create project: %w", err)
	}
	return nil
}

func UpdateProject(db Writer, p *core.Project) error {
	res, err := db.Exec(`
		UPDATE projects SET
			name = ?, enabled = ?, upstream = ?, algorithm = ?,
			fw_limit = ?, fw_window = ?,
			sw_limit = ?, sw_window = ?,
			tb_capacity = ?, tb_refill_rate = ?,
			updated_at = unixepoch()
		WHERE id = ?`,
		p.Name, p.Enabled, p.Upstream, p.Algorithm,
		nullInt(p.FWLimit), nullDuration(p.FWWindow),
		nullInt(p.SWLimit), nullDuration(p.SWWindow),
		nullInt(p.TBCapacity), p.TBRefillRate,
		p.ID,
	)
	if err != nil {
		return fmt.Errorf("update project: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func DeleteProject(db Writer, id string) error {
	res, err := db.Exec(`DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// helpers

func nullInt(v *int) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*v), Valid: true}
}

func nullDuration(v *time.Duration) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*v), Valid: true}
}
