package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProductsPageHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/products", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("expected HTML content type, got %v", ct)
	}
}

func TestProductsPageFilterBySource(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/products?source=fptshop", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "fptshop") {
		t.Errorf("expected body to contain fptshop products")
	}
}

func TestProductsPageFilterByCategory(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/products?category=Laptop", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Laptop") {
		t.Errorf("expected body to contain Laptop products")
	}
}

func TestProductsPagePagination(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/products?page=1", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
	}
}

func TestProductsPageInvalidPage(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/products?page=invalid", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
	}
}

func TestProductsHXRequest(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/products?source=fptshop", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("expected HTML content type, got %v", ct)
	}
}
