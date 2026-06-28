package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/useopenward/openward/internal/core"
	"github.com/useopenward/openward/internal/db"
)

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := db.ListProjects(s.db.Reader)
	if err != nil {
		http.Error(w, "failed to list projects", http.StatusInternalServerError)
		return
	}
	writeJSON(w, projects)
}

func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := db.GetProject(s.db.Reader, id)
	if errors.Is(err, db.ErrNotFound) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to get project", http.StatusInternalServerError)
		return
	}
	writeJSON(w, p)
}

func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name      string `json:"name"`
		Upstream  string `json:"upstream"`
		Algorithm core.RateLimitAlgorithm `json:"algorithm"`
		Enabled   bool                    `json:"enabled"`

		FWLimit  *int           `json:"fw_limit"`
		FWWindow *time.Duration `json:"fw_window"`

		SWLimit  *int           `json:"sw_limit"`
		SWWindow *time.Duration `json:"sw_window"`

		TBCapacity   *int     `json:"tb_capacity"`
		TBRefillRate *float64 `json:"tb_refill_rate"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.Upstream == "" || req.Algorithm == "" {
		http.Error(w, "name, upstream, and algorithm are required", http.StatusBadRequest)
		return
	}
	if !isValidRateLimitAlgorithm(req.Algorithm) {
		http.Error(w, "invalid algorithm", http.StatusBadRequest)
		return
	}

	p := &core.Project{
		ID:           uuid.NewString(),
		Name:         req.Name,
		APIKey:       uuid.NewString(),
		Enabled:      req.Enabled,
		Upstream:     req.Upstream,
		Algorithm:    req.Algorithm,
		FWLimit:      req.FWLimit,
		FWWindow:     req.FWWindow,
		SWLimit:      req.SWLimit,
		SWWindow:     req.SWWindow,
		TBCapacity:   req.TBCapacity,
		TBRefillRate: req.TBRefillRate,
	}

	if err := db.CreateProject(s.db.Writer, p); err != nil {
		http.Error(w, "failed to create project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	writeJSON(w, p)
}

func (s *Server) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	existing, err := db.GetProject(s.db.Reader, id)
	if errors.Is(err, db.ErrNotFound) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to get project", http.StatusInternalServerError)
		return
	}

	// decode only provided fields onto the existing project
	var req struct {
		Name      *string `json:"name"`
		Upstream  *string `json:"upstream"`
		Algorithm *core.RateLimitAlgorithm `json:"algorithm"`
		Enabled   *bool                    `json:"enabled"`

		FWLimit  *int           `json:"fw_limit"`
		FWWindow *time.Duration `json:"fw_window"`

		SWLimit  *int           `json:"sw_limit"`
		SWWindow *time.Duration `json:"sw_window"`

		TBCapacity   *int     `json:"tb_capacity"`
		TBRefillRate *float64 `json:"tb_refill_rate"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Upstream != nil {
		existing.Upstream = *req.Upstream
	}
	if req.Algorithm != nil {
		if !isValidRateLimitAlgorithm(*req.Algorithm) {
			http.Error(w, "invalid algorithm", http.StatusBadRequest)
			return
		}
		existing.Algorithm = *req.Algorithm
	}
	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}
	if req.FWLimit != nil {
		existing.FWLimit = req.FWLimit
	}
	if req.FWWindow != nil {
		existing.FWWindow = req.FWWindow
	}
	if req.SWLimit != nil {
		existing.SWLimit = req.SWLimit
	}
	if req.SWWindow != nil {
		existing.SWWindow = req.SWWindow
	}
	if req.TBCapacity != nil {
		existing.TBCapacity = req.TBCapacity
	}
	if req.TBRefillRate != nil {
		existing.TBRefillRate = req.TBRefillRate
	}

	if err := db.UpdateProject(s.db.Writer, existing); err != nil {
		http.Error(w, "failed to update project", http.StatusInternalServerError)
		return
	}

	writeJSON(w, existing)
}

func (s *Server) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := db.DeleteProject(s.db.Writer, id)
	if errors.Is(err, db.ErrNotFound) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to delete project", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func isValidRateLimitAlgorithm(a core.RateLimitAlgorithm) bool {
	switch a {
	case core.AlgoFixedWindow, core.AlgoSlidingWindow, core.AlgoTokenBucket:
		return true
	default:
		return false
	}
}
