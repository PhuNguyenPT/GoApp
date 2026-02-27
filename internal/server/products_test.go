package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
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

func TestProductDetailPageHandler(t *testing.T) {
	validID, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("failed to generate uuid: %v", err)
	}
	t.Run("valid id returns product page", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/products/"+validID.String(), nil)
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
		if !strings.Contains(rr.Body.String(), "iPhone 17 Pro") {
			t.Errorf("expected body to contain product name")
		}
	})

	t.Run("invalid id redirects to products", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/products/not-a-valid-uuid", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusFound {
			t.Errorf("expected redirect 302, got %v", rr.Code)
		}
		if rr.Header().Get("Location") != "/products" {
			t.Errorf("expected redirect to /products, got %v", rr.Header().Get("Location"))
		}
	})
}
