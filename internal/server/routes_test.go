package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type mockDB struct {}

func (m *mockDB) Health() map[string]string {
	return map[string]string{ "status": "up", "message": "It's healthy" }
}

func (m *mockDB) Close() error  {
	return nil
}

var testHandler http.Handler
func TestMain(m *testing.M)  {
	s := &Server{db: &mockDB{}}
	testHandler = s.RegisterRoutes(&Config{GinMode: gin.TestMode})
	m.Run()
}

func TestApiInfoHandler(t *testing.T)  {
	req, _ := http.NewRequest(http.MethodGet, "/api/", nil)
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

func TestHealthHandler(t *testing.T)  {
	req, _ := http.NewRequest(http.MethodGet, "/api/health", nil)
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

func TestHomePageHandler(t *testing.T)  {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
	}

	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("expected HTML content type, got %v", ct)
	}
}

func TestContactFormHandler(t *testing.T) {
	form := url.Values{}
	form.Set("name", "Test Name")
	form.Set("email", "test@example.com")
	form.Set("subject", "Test Subject")
	form.Set("message", "Test message")

	req, _ := http.NewRequest(http.MethodPost, "/contact", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "Test Nam") {
		t.Errorf("expected response body to contain 'Test Name'")
	}
}

func TestContactPageHandler(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/contact", nil)
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("expected HTML content type, got %v", ct)
	}
}

func TestUnknownRoute(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/does-not-exist", nil)
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %v", rr.Code)
	}
}
