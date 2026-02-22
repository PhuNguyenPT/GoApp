package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApiInfoHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/api/", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON %v", err)
	}

	if body["message"] != "Go Server API" {
		t.Errorf("unexpected message: %v", body["message"])
	}

	if body["version"] != "0.0" {
		t.Errorf("unexpected version: %v", body["version"])
	}
}

func TestHealthHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/api/health", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	if body["status"] != "up" {
		t.Errorf("expected status up, got %v", body["status"])
	}

	if body["message"] != "It's healthy" {
		t.Errorf("expected healthy message, got %v", body["message"])
	}
}
