// Package server exposes the cache-scanning engine over a small
// JSON+SSE HTTP API, bound to loopback only, for the embedded
// browser-based UI.
package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DavidMarsanic/disk-space-cleaner/internal/engine"
	"github.com/DavidMarsanic/disk-space-cleaner/internal/jobs"
	"github.com/DavidMarsanic/disk-space-cleaner/web"
)

const idleTimeout = 30 * time.Minute

type Server struct {
	Jobs *jobs.Registry
	ctx  context.Context

	mu      sync.Mutex
	results map[string][]engine.Category // jobID -> its finished scan result

	lastActivity atomic.Int64
}

func New(ctx context.Context) *Server {
	s := &Server{
		ctx:     ctx,
		Jobs:    jobs.NewRegistry(),
		results: map[string][]engine.Category{},
	}
	s.touch()
	return s
}

func (s *Server) storeResult(jobID string, categories []engine.Category) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results[jobID] = categories
}

func (s *Server) getResult(jobID string) ([]engine.Category, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.results[jobID]
	return r, ok
}

// Start binds 127.0.0.1:port (port 0 picks any free port — this UI is
// never exposed beyond loopback) and serves until the process exits.
func (s *Server) Start(port int) (string, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return "", fmt.Errorf("starting local server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/scan", s.handleScan)
	mux.HandleFunc("GET /api/jobs/{id}/events", s.handleJobEvents)
	mux.HandleFunc("GET /api/scans/{id}", s.handleScanResult)
	mux.HandleFunc("POST /api/clean", s.handleClean)
	mux.HandleFunc("GET /api/trash-info", s.handleTrashInfo)
	mux.HandleFunc("POST /api/trash-empty", s.handleTrashEmpty)
	mux.HandleFunc("POST /api/reveal", s.handleReveal)
	mux.HandleFunc("POST /api/open", s.handleOpen)
	mux.Handle("GET /", http.FileServer(http.FS(web.Static)))

	httpSrv := &http.Server{Handler: s.trackActivity(mux)}
	go func() {
		_ = httpSrv.Serve(ln)
	}()
	go s.watchIdle()

	return "http://" + ln.Addr().String(), nil
}

func (s *Server) trackActivity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.touch()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) touch() {
	s.lastActivity.Store(time.Now().Unix())
}

func (s *Server) watchIdle() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		idleFor := time.Now().Unix() - s.lastActivity.Load()
		if idleFor > int64(idleTimeout.Seconds()) && !s.Jobs.HasActive() {
			os.Exit(0)
		}
	}
}
