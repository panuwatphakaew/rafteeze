//TODO: Use Gin
package http

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type ProposeFunc func(key, value string) error
type GetFunc func(key string) (string, bool)

type Server struct {
	addr    string
	propose ProposeFunc
	get     GetFunc
}

func NewServer(addr string, propose ProposeFunc, get GetFunc) *Server {
	return &Server{
		addr:    addr,
		propose: propose,
		get:     get,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/kv/", s.handleKV)

	slog.Info("HTTP server starting", slog.String("addr", s.addr))
	return http.ListenAndServe(s.addr, mux)
}

func (s *Server) handleKV(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/kv/")
	if key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPut, http.MethodPost:
		s.handlePut(w, r, key)
	case http.MethodGet:
		s.handleGet(w, r, key)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePut(w http.ResponseWriter, r *http.Request, key string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("failed to read request body", slog.Any("error", err))
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	value := string(body)

	if err := s.propose(key, value); err != nil {
		slog.Error("failed to propose", slog.Any("error", err))
		http.Error(w, fmt.Sprintf("Failed to propose: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "proposed",
		"key":    key,
	})
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request, key string) {
	value, exists := s.get(key)
	if !exists {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(value))
}

