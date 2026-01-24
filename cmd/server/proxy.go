package main

import (
	"io"
	"log"
	"net/http"
	"time"
)

var proxyClient = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 1000,
		IdleConnTimeout:     90 * time.Second,
	},
}

func forwardToLeader(w http.ResponseWriter, r *http.Request, leaderURL string) {
	req, err := http.NewRequest(r.Method, leaderURL, r.Body)
	if err != nil {
		http.Error(w, "Failed to create forwarding request", http.StatusInternalServerError)
		return
	}

	resp, err := proxyClient.Do(req)
	if err != nil {
		log.Printf("Proxy error: %v", err)
		http.Error(w, "Failed to forward request to leader", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)

	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("Error copying response: %v", err)
	}
}
