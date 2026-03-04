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
	body := rr.Body.String()
	// Should show all products and both sources in filter
	if !strings.Contains(body, "fptshop") {
		t.Errorf("expected body to contain source fptshop")
	}
	if !strings.Contains(body, "thegioididong") {
		t.Errorf("expected body to contain source thegioididong")
	}
	// Should show filter panel
	if !strings.Contains(body, "source-filter") {
		t.Errorf("expected body to contain source-filter panel")
	}
}

func TestProductsPageFilterBySource(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		wantContains []string
		wantAbsent   []string
	}{
		{
			name:         "fptshop source shows only fptshop products",
			source:       "fptshop",
			wantContains: []string{"iPhone 17 Pro", "Acer Predator"},
			// Use a product name unique to thegioididong that won't appear in filter links
			wantAbsent: []string{"Laptop Acer Gaming Nitro", "iPhone 17 Pro Max 256GB"},
		},
		{
			name:         "thegioididong source shows only thegioididong products",
			source:       "thegioididong",
			wantContains: []string{"iPhone 17 Pro Max 256GB", "Laptop Acer Gaming Nitro"},
			// Use full product names unique to fptshop
			wantAbsent: []string{"Acer Predator Helios 18 Gaming AI", "iPhone 17 Pro&#34;"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/products?source="+tt.source, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			rr := httptest.NewRecorder()
			testHandler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
			}
			body := rr.Body.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(body, want) {
					t.Errorf("expected body to contain %q", want)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(body, absent) {
					t.Errorf("expected body NOT to contain %q", absent)
				}
			}
		})
	}
}

func TestProductsPageFilterByCategory(t *testing.T) {
	tests := []struct {
		name         string
		category     string
		wantContains []string
		wantAbsent   []string
	}{
		{
			name:         "Laptop category shows only laptops",
			category:     "Laptop",
			wantContains: []string{"Laptop", "Acer"},
			wantAbsent:   []string{"iPhone"},
		},
		{
			name:         "Phone category shows only phones",
			category:     "Điện thoại",
			wantContains: []string{"iPhone"},
			wantAbsent:   []string{"Acer Predator"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/products?category="+tt.category, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			rr := httptest.NewRecorder()
			testHandler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
			}
			body := rr.Body.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(body, want) {
					t.Errorf("expected body to contain %q", want)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(body, absent) {
					t.Errorf("expected body NOT to contain %q", absent)
				}
			}
		})
	}
}

func TestProductsPageFilterBySubcategory(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		category     string
		subcategory  string
		wantContains []string
		wantAbsent   []string
	}{
		{
			name:         "fptshop Laptop Acer subcategory",
			source:       "fptshop",
			category:     "Laptop",
			subcategory:  "Acer",
			wantContains: []string{"Acer Predator"},
			wantAbsent:   []string{"iPhone"},
		},
		{
			name:         "thegioididong phone Apple subcategory",
			source:       "thegioididong",
			category:     "Điện thoại",
			subcategory:  "Điện thoại iPhone (Apple)",
			wantContains: []string{"iPhone 17 Pro Max"},
			wantAbsent:   []string{"Acer"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/products?source=" + tt.source + "&category=" + tt.category + "&subcategory=" + tt.subcategory
			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			rr := httptest.NewRecorder()
			testHandler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
			}
			body := rr.Body.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(body, want) {
					t.Errorf("expected body to contain %q", want)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(body, absent) {
					t.Errorf("expected body NOT to contain %q", absent)
				}
			}
		})
	}
}

func TestProductsPagePagination(t *testing.T) {
	tests := []struct {
		name       string
		page       string
		wantStatus int
	}{
		{"valid page 1", "1", http.StatusOK},
		{"valid page 2", "2", http.StatusOK},
		{"invalid page string", "invalid", http.StatusOK}, // falls back to page 1
		{"zero page", "0", http.StatusOK},                 // falls back to page 1
		{"negative page", "-1", http.StatusOK},            // falls back to page 1
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/products?page="+tt.page, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			rr := httptest.NewRecorder()
			testHandler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected status %v, got %v", tt.wantStatus, rr.Code)
			}
		})
	}
}

func TestProductsHXRequest(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		trigger      string
		wantContains []string
		wantAbsent   []string
	}{
		{
			name:         "source trigger clears category and returns OOB + grid",
			url:          "/products?source=fptshop&trigger=source",
			trigger:      "source",
			wantContains: []string{"source-filter", "product-results", "active-chips"},
		},
		{
			name:         "category trigger returns subcategory OOB + grid",
			url:          "/products?source=fptshop&category=Laptop&trigger=category",
			trigger:      "category",
			wantContains: []string{"subcategory-filter", "active-chips"},
		},
		{
			name:         "no trigger returns grid only with OOBs",
			url:          "/products?source=fptshop&category=Laptop",
			wantContains: []string{"source-filter", "active-chips"},
		},
		{
			name:    "source trigger resets category filter",
			url:     "/products?source=fptshop&category=Laptop&trigger=source",
			trigger: "source",
			// category should be cleared server-side so Laptop products won't be filtered
			wantContains: []string{"fptshop"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tt.url, nil)
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
			body := rr.Body.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(body, want) {
					t.Errorf("expected body to contain %q", want)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(body, absent) {
					t.Errorf("expected body NOT to contain %q", absent)
				}
			}
		})
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
		body := rr.Body.String()
		if !strings.Contains(body, "iPhone 17 Pro") {
			t.Errorf("expected body to contain product name")
		}
		// Should show product details elements
		if !strings.Contains(body, "fptshop") {
			t.Errorf("expected body to contain product source")
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
