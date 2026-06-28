package api

import (
	"net/http"
	"time"

	"github.com/useopenward/openward/internal/db"
)

type Server struct {
	db  *db.Handles
	mux *http.ServeMux
}

func NewServer(database *db.Handles) *http.Server {
	s := &Server{
		db:  database,
		mux: http.NewServeMux(),
	}
	s.routes()
	return &http.Server{
		Addr:         ":9090",
		Handler:      s.mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func (s *Server) routes() {
	// public
	s.mux.HandleFunc("POST /api/auth/login", s.handleLogin)
	s.mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// protected
	s.mux.HandleFunc("GET /api/projects", s.auth(s.handleListProjects))
	s.mux.HandleFunc("POST /api/projects", s.auth(s.handleCreateProject))
	s.mux.HandleFunc("GET /api/projects/{id}", s.auth(s.handleGetProject))
	s.mux.HandleFunc("PATCH /api/projects/{id}", s.auth(s.handleUpdateProject))
	s.mux.HandleFunc("DELETE /api/projects/{id}", s.auth(s.handleDeleteProject))
	s.mux.HandleFunc("GET /api/projects/{id}/logs", s.auth(s.handleListLogs))

	// SPA — catch-all, must be last
	s.mux.Handle("/", staticHandler())
}
