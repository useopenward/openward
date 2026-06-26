package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/useopenward/openward/internal/core"
	dbpkg "github.com/useopenward/openward/internal/db"
	"github.com/useopenward/openward/internal/limiter"
)

const cacheTTL = 30 * time.Second

type cachedProject struct {
	project   *core.Project
	expiresAt time.Time
}

type Handler struct {
	reader dbpkg.Reader
	writer dbpkg.Writer
	cache  sync.Map // map[apiKey]*cachedProject
}

func NewHandler(reader dbpkg.Reader, writer dbpkg.Writer) *Handler {
	return &Handler{reader: reader, writer: writer}
}

func (h *Handler) getProject(apiKey string) (*core.Project, error) {
	if v, ok := h.cache.Load(apiKey); ok {
		entry := v.(*cachedProject)
		if time.Now().Before(entry.expiresAt) {
			return entry.project, nil
		}
		h.cache.Delete(apiKey) // expired
	}

	p, err := dbpkg.GetProjectByAPIKey(h.reader, apiKey)
	if err != nil {
		return nil, err
	}

	h.cache.Store(apiKey, &cachedProject{
		project:   p,
		expiresAt: time.Now().Add(cacheTTL),
	})
	return p, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. authenticate
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		http.Error(w, "missing api key", http.StatusUnauthorized)
		return
	}

	project, err := h.getProject(apiKey)
	if err != nil {
		http.Error(w, "invalid api key", http.StatusUnauthorized)
		return
	}

	if !project.Enabled {
		http.Error(w, "project disabled", http.StatusForbidden)
		return
	}

	// 2. rate limit
	l, err := limiter.New(project)
	if err != nil {
		log.Printf("limiter init error for project %s: %v", project.ID, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	allowed, err := l.Allow(h.reader, h.writer, project.ID)
	if err != nil {
		log.Printf("limiter error for project %s: %v", project.ID, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if !allowed {
		h.logRequest(project.ID, false, 0)
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rateLimit(project)))
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	// 3. proxy
	upstream, err := url.Parse(project.Upstream)
	if err != nil {
		log.Printf("invalid upstream for project %s: %v", project.ID, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = upstream.Scheme
			req.URL.Host = upstream.Host
			req.Host = upstream.Host
			req.Header.Del("X-API-Key")
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("proxy error for project %s: %v", project.ID, err)
			recorder.status = http.StatusBadGateway
			http.Error(w, "upstream error", http.StatusBadGateway)
		},
	}

	proxy.ServeHTTP(recorder, r)
	go h.logRequest(project.ID, true, recorder.status)
}

func (h *Handler) logRequest(projectID string, allowed bool, statusCode int) {
	var sc *int
	if statusCode != 0 {
		sc = &statusCode
	}
	_, err := h.writer.Exec(`
		INSERT INTO request_logs (project_id, requested_at, allowed, status_code)
		VALUES (?, unixepoch(), ?, ?)
	`, projectID, allowed, sc)
	if err != nil {
		log.Printf("failed to log request for project %s: %v", projectID, err)
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func rateLimit(p *core.Project) int {
	switch p.Algorithm {
	case core.AlgoFixedWindow:
		if p.FWLimit != nil {
			return *p.FWLimit
		}
	case core.AlgoSlidingWindow:
		if p.SWLimit != nil {
			return *p.SWLimit
		}
	case core.AlgoTokenBucket:
		if p.TBCapacity != nil {
			return *p.TBCapacity
		}
	}
	return 0
}
