package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body["status"] != "healthy" {
		t.Errorf("expected healthy, got %v", body["status"])
	}
	if body["service"] != "collector" {
		t.Errorf("expected collector, got %v", body["service"])
	}
}

func TestMetricsPostHandler(t *testing.T) {
	store = NewMetricsStore()

	payload := map[string]interface{}{"cpu": 65.2, "memory": 1024}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/metrics", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	metricsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["received"] != true {
		t.Errorf("expected received=true")
	}
	if resp["count"] != float64(2) {
		t.Errorf("expected count=2, got %v", resp["count"])
	}

	if store.Count() != 2 {
		t.Errorf("store should have 2 entries, got %d", store.Count())
	}
}

func TestMetricsPostInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/metrics", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	metricsHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestMetricsGetHandler(t *testing.T) {
	store = NewMetricsStore()
	store.Add("test_metric", 42)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	metricsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	metrics, ok := resp["metrics"].([]interface{})
	if !ok {
		t.Fatalf("expected metrics array")
	}
	if len(metrics) != 1 {
		t.Errorf("expected 1 metric, got %d", len(metrics))
	}
}

func TestMetricsStoreCap(t *testing.T) {
	s := NewMetricsStore()
	for i := 0; i < 1050; i++ {
		s.Add("m", i)
	}
	if s.Count() != 1000 {
		t.Errorf("expected store capped at 1000, got %d", s.Count())
	}
}

func TestMetricsMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/metrics", nil)
	w := httptest.NewRecorder()
	metricsHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
