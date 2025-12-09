package main

import (
	"fmt"
	"net/http"
)

type Server struct {
	store *KVStore
}

func NewServer() *Server {
	return &Server{store: NewKVStore()}
}

func (s *Server) handlePut(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	val := r.URL.Query().Get("val")

	if key == "" || val == "" {
		http.Error(w, "Missing key/val", http.StatusBadRequest)
		return
	}
	s.store.Put(key, val)
	_, err := fmt.Fprintf(w, "Success Put: %s in %s", key, val)
	if err != nil {
		return
	}
}
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "No key found", http.StatusBadRequest)
		return
	}
	val, ok := s.store.Get(key)

	if !ok {
		http.Error(w, "No key found", http.StatusBadRequest)
		return
	}
	_, err := fmt.Fprintf(w, "Success Get: %s -> %s", key, val)
	if err != nil {
		return
	}
}
