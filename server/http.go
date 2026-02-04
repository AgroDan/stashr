package server

import (
	"encoding/json"
	"net/http"
	"time"

	"kvstore/store"
)

type HTTPServer struct {
	store *store.Store
	mux   *http.ServeMux
}

func NewHTTPServer(s *store.Store) *HTTPServer {
	h := &HTTPServer{store: s, mux: http.NewServeMux()}
	h.mux.HandleFunc("GET /keys/{key}", h.handleGet)
	h.mux.HandleFunc("PUT /keys/{key}", h.handleSet)
	h.mux.HandleFunc("DELETE /keys/{key}", h.handleDelete)
	return h
}

func (h *HTTPServer) Handler() http.Handler {
	return h.mux
}

func (h *HTTPServer) handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	val, ok := h.store.Get(key)
	if !ok {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"value": val})
}

type setRequest struct {
	Value      string `json:"value"`
	TTLSeconds int64  `json:"ttl_seconds"`
}

func (h *HTTPServer) handleSet(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	var req setRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	var ttl time.Duration
	if req.TTLSeconds > 0 {
		ttl = time.Duration(req.TTLSeconds) * time.Second
	}

	h.store.Set(key, req.Value, ttl)
	w.WriteHeader(http.StatusNoContent)
}

func (h *HTTPServer) handleDelete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	deleted := h.store.Delete(key)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"deleted": deleted})
}
