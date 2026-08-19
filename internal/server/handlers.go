package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DavidMarsanic/disk-space-cleaner/internal/browser"
	"github.com/DavidMarsanic/disk-space-cleaner/internal/engine"
	"github.com/DavidMarsanic/disk-space-cleaner/internal/jobs"
	"github.com/DavidMarsanic/disk-space-cleaner/internal/trash"
)

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	job, ctx := s.Jobs.Create(s.ctx)

	go func() {
		onProgress := func(p engine.Progress) {
			job.Publish(jobs.Event{Stage: p.Stage, Message: p.Message})
		}
		categories, err := engine.Scan(ctx, onProgress)
		if err != nil {
			if ctx.Err() != nil {
				job.Publish(jobs.Event{Stage: "canceled"})
				return
			}
			job.Publish(jobs.Event{Stage: "error", Message: err.Error()})
			return
		}
		s.storeResult(job.ID, categories)
		job.Publish(jobs.Event{Stage: "done"})
	}()

	writeJSON(w, http.StatusOK, map[string]string{"jobId": job.ID})
}

func (s *Server) handleScanResult(w http.ResponseWriter, r *http.Request) {
	result, ok := s.getResult(r.PathValue("id"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"categories": result})
}

type cleanOutcome struct {
	ID    string `json:"id"`
	Error string `json:"error,omitempty"`
}

// handleClean moves every path belonging to the requested categories to
// the OS Trash — never an arbitrary client-supplied path, only the exact
// paths this server itself recorded for jobId's scan. Best-effort per
// category: one failure (permissions, already moved) doesn't block the
// rest of the batch.
func (s *Server) handleClean(w http.ResponseWriter, r *http.Request) {
	var req struct {
		JobID string   `json:"jobId"`
		IDs   []string `json:"ids"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	categories, ok := s.getResult(req.JobID)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unknown scan — run a new scan", "code": "bad-request"})
		return
	}
	byID := map[string]engine.Category{}
	for _, c := range categories {
		byID[c.ID] = c
	}

	results := make([]cleanOutcome, 0, len(req.IDs))
	for _, id := range req.IDs {
		cat, ok := byID[id]
		if !ok {
			results = append(results, cleanOutcome{ID: id, Error: "unknown category in this scan"})
			continue
		}
		var firstErr error
		for _, p := range cat.Paths {
			if err := trash.Move(p); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		if firstErr != nil {
			results = append(results, cleanOutcome{ID: id, Error: firstErr.Error()})
			continue
		}
		results = append(results, cleanOutcome{ID: id})
	}

	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

func (s *Server) handleTrashInfo(w http.ResponseWriter, r *http.Request) {
	bytes, items, err := trash.Info()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"bytes": bytes, "items": items})
}

// handleTrashEmpty permanently deletes everything in the OS Trash — the
// one genuinely irreversible action in this whole app. The frontend
// requires its own explicit confirmation before ever calling this.
func (s *Server) handleTrashEmpty(w http.ResponseWriter, r *http.Request) {
	if err := trash.Empty(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleJobEvents(w http.ResponseWriter, r *http.Request) {
	job, ok := s.Jobs.Get(r.PathValue("id"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	ch, cancel := job.Subscribe()
	defer cancel()

	for {
		select {
		case e, open := <-ch:
			if !open {
				return
			}
			data, _ := json.Marshal(e)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
			if e.Stage == "done" || e.Stage == "error" || e.Stage == "canceled" {
				return
			}
		case <-r.Context().Done():
			return
		}
	}
}

func (s *Server) handleReveal(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := browser.Reveal(req.Path); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleOpen(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := browser.Open(req.Path); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body", "code": "bad-request"})
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
