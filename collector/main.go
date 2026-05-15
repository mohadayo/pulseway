package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type MetricEntry struct {
	Name      string      `json:"name"`
	Value     interface{} `json:"value"`
	Timestamp time.Time   `json:"timestamp"`
}

type MetricsStore struct {
	mu      sync.RWMutex
	entries []MetricEntry
}

func NewMetricsStore() *MetricsStore {
	return &MetricsStore{entries: make([]MetricEntry, 0)}
}

func (s *MetricsStore) Add(name string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, MetricEntry{
		Name:      name,
		Value:     value,
		Timestamp: time.Now().UTC(),
	})
	if len(s.entries) > 1000 {
		s.entries = s.entries[len(s.entries)-1000:]
	}
}

func (s *MetricsStore) GetAll() []MetricEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]MetricEntry, len(s.entries))
	copy(result, s.entries)
	return result
}

func (s *MetricsStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}

var (
	store     = NewMetricsStore()
	startTime = time.Now()
	logger    = log.New(os.Stdout, "[collector] ", log.LstdFlags)
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         "healthy",
		"service":        "collector",
		"uptime_seconds": time.Since(startTime).Seconds(),
		"metrics_count":  store.Count(),
	})
}

func metricsPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		logger.Printf("ERROR: invalid JSON body: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON body"})
		return
	}

	for name, value := range body {
		store.Add(name, value)
		logger.Printf("INFO: recorded metric %s = %v", name, value)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"received": true, "count": len(body)})
}

func metricsGetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}
	logger.Println("INFO: serving all metrics")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"metrics": store.GetAll()})
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		metricsGetHandler(w, r)
	case http.MethodPost:
		metricsPostHandler(w, r)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
	}
}

func main() {
	port := os.Getenv("COLLECTOR_PORT")
	if port == "" {
		port = "8001"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/metrics", metricsHandler)

	addr := fmt.Sprintf("0.0.0.0:%s", port)
	logger.Printf("INFO: starting collector on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Fatalf("FATAL: server failed: %v", err)
	}
}
