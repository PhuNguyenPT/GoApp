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
func TestProductsPageSorting(t *testing.T) {
	tests := []struct {
		name         string
		sort         string
		wantStatus   int
		wantContains []string
	}{
		{
			name:         "sort by price asc",
			sort:         "price,asc",
			wantStatus:   http.StatusOK,
			wantContains: []string{"product-results"},
		},
		{
			name:         "sort by price desc",
			sort:         "price,desc",
			wantStatus:   http.StatusOK,
			wantContains: []string{"product-results"},
		},
		{
			name:         "sort by rating desc",
			sort:         "rating,desc",
			wantStatus:   http.StatusOK,
			wantContains: []string{"product-results"},
		},
		{
			name:         "sort by name asc",
			sort:         "name,asc",
			wantStatus:   http.StatusOK,
			wantContains: []string{"product-results"},
		},
		{
			name:       "invalid sort field falls back to default",
			sort:       "invalid_field,asc",
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing sort returns default order",
			sort:       "",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/products"
			if tt.sort != "" {
				url += "?sort=" + tt.sort
			}
			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			rr := httptest.NewRecorder()
			testHandler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected status %v, got %v", tt.wantStatus, rr.Code)
			}
			body := rr.Body.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(body, want) {
					t.Errorf("expected body to contain %q", want)
				}
			}
		})
	}
}

func TestProductsPagePriceFilter(t *testing.T) {
	tests := []struct {
		name       string
		minPrice   string
		maxPrice   string
		wantStatus int
	}{
		{"valid price range", "100000", "5000000", http.StatusOK},
		{"zero min and max (all prices)", "0", "0", http.StatusOK},
		{"only min price set", "500000", "0", http.StatusOK},
		{"only max price set", "0", "50000000", http.StatusOK},
		{"negative min falls back to zero", "-100", "1000000", http.StatusOK},
		{"negative max falls back to zero", "0", "-500", http.StatusOK},
		{"non-numeric min falls back to zero", "abc", "1000000", http.StatusOK},
		{"non-numeric max falls back to zero", "0", "xyz", http.StatusOK},
		{"min greater than max still returns 200", "9000000", "1000000", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/products?min_price=" + tt.minPrice + "&max_price=" + tt.maxPrice
			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			rr := httptest.NewRecorder()
			testHandler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected status %v, got %v", tt.wantStatus, rr.Code)
			}
			if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
				t.Errorf("expected HTML content type, got %v", ct)
			}
		})
	}
}

func TestProductsFragmentHandler(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		wantStatus   int
		wantContains []string
	}{
		{
			name:         "no params returns default products",
			url:          "/products-fragment",
			wantStatus:   http.StatusOK,
			wantContains: []string{"iPhone", "Acer"},
		},
		{
			name:         "filter by source fptshop",
			url:          "/products-fragment?source=fptshop",
			wantStatus:   http.StatusOK,
			wantContains: []string{"fptshop"},
		},
		{
			name:         "filter by category Laptop",
			url:          "/products-fragment?category=Laptop",
			wantStatus:   http.StatusOK,
			wantContains: []string{"Acer"},
		},
		{
			name:       "valid custom limit",
			url:        "/products-fragment?limit=2",
			wantStatus: http.StatusOK,
		},
		{
			name:       "limit zero falls back to default",
			url:        "/products-fragment?limit=0",
			wantStatus: http.StatusOK,
		},
		{
			name:       "limit above 100 falls back to default",
			url:        "/products-fragment?limit=999",
			wantStatus: http.StatusOK,
		},
		{
			name:       "negative limit falls back to default",
			url:        "/products-fragment?limit=-5",
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-numeric limit falls back to default",
			url:        "/products-fragment?limit=abc",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tt.url, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			rr := httptest.NewRecorder()
			testHandler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected status %v, got %v", tt.wantStatus, rr.Code)
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
		})
	}
}

func TestProductsHXRequestWithPriceAndSort(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		wantStatus   int
		wantContains []string
	}{
		{
			name:         "htmx with price range returns OOBs and grid",
			url:          "/products?min_price=1000000&max_price=50000000",
			wantStatus:   http.StatusOK,
			wantContains: []string{"source-filter", "active-chips", "product-results"},
		},
		{
			name:         "htmx with sort returns grid",
			url:          "/products?sort=price,asc",
			wantStatus:   http.StatusOK,
			wantContains: []string{"source-filter", "product-results"},
		},
		{
			name:         "htmx source + price filter",
			url:          "/products?source=thegioididong&min_price=100000&max_price=2000000&trigger=source",
			wantStatus:   http.StatusOK,
			wantContains: []string{"source-filter", "active-chips"},
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

			if rr.Code != tt.wantStatus {
				t.Errorf("expected status %v, got %v", tt.wantStatus, rr.Code)
			}
			body := rr.Body.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(body, want) {
					t.Errorf("expected body to contain %q", want)
				}
			}
		})
	}
}

func TestProductsPageCombinedFiltersAndSort(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		wantContains []string
		wantAbsent   []string
	}{
		{
			name:         "fptshop laptops sorted by price asc",
			url:          "/products?source=fptshop&category=Laptop&sort=price,asc",
			wantContains: []string{"Acer Predator"},
			wantAbsent:   []string{"iPhone"},
		},
		{
			name:         "thegioididong phones sorted by rating desc",
			url:          "/products?source=thegioididong&category=Điện thoại&sort=rating,desc",
			wantContains: []string{"iPhone 17 Pro Max"},
			wantAbsent:   []string{"Acer Predator"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tt.url, nil)
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

func TestProductsPageFilterCombinations(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		wantContains []string
		wantAbsent   []string
	}{
		{
			name:         "subcategory alone shows matching products",
			url:          "/products?subcategory=Apple",
			wantContains: []string{"iPhone 17 Pro"},
			wantAbsent:   []string{"Acer Predator", "Laptop Acer Gaming Nitro"},
		},
		{
			name:         "source + price range filters correctly",
			url:          "/products?source=thegioididong&min_price=100000&max_price=2000000",
			wantContains: []string{"thegioididong"},
			wantAbsent:   []string{"Acer Predator Helios 18 Gaming AI"},
		},
		{
			name:         "source + sort by name asc",
			url:          "/products?source=thegioididong&sort=name,asc",
			wantContains: []string{"thegioididong"},
			wantAbsent:   []string{"Acer Predator Helios 18 Gaming AI"},
		},
		{
			name:         "category + price range filters correctly",
			url:          "/products?category=Laptop&min_price=100000&max_price=50000000",
			wantContains: []string{"Acer"},
			wantAbsent:   []string{"iPhone"},
		},
		{
			name:         "category + sort by price desc",
			url:          "/products?category=Laptop&sort=price,desc",
			wantContains: []string{"Acer"},
			wantAbsent:   []string{"iPhone"},
		},
		{
			name:         "price + sort returns all products in order",
			url:          "/products?min_price=100000&max_price=200000000&sort=price,asc",
			wantContains: []string{"source-filter"},
		},
		{
			name:         "source + category + price",
			url:          "/products?source=fptshop&category=Laptop&min_price=100000&max_price=200000000",
			wantContains: []string{"Acer Predator"},
			wantAbsent:   []string{"iPhone", "Laptop Acer Gaming Nitro"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tt.url, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			rr := httptest.NewRecorder()
			testHandler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("expected status 200, got %v", rr.Code)
			}
			if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
				t.Errorf("expected text/html, got %q", ct)
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

func TestProductsPageAllFilters(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		wantContains []string
		wantAbsent   []string
	}{
		{
			name:         "all five: fptshop + Laptop + Acer + price + sort by price asc",
			url:          "/products?source=fptshop&category=Laptop&subcategory=Acer&min_price=100000&max_price=200000000&sort=price,asc",
			wantContains: []string{"Acer Predator"},
			wantAbsent:   []string{"iPhone", "Laptop Acer Gaming Nitro"},
		},
		{
			name:         "all five: thegioididong + phone + Apple subcategory + price + sort by rating desc",
			url:          "/products?source=thegioididong&category=Điện thoại&subcategory=Điện thoại iPhone (Apple)&min_price=100000&max_price=200000000&sort=rating,desc",
			wantContains: []string{"iPhone 17 Pro Max"},
			wantAbsent:   []string{"Acer Predator", "Laptop Acer Gaming Nitro"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tt.url, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			rr := httptest.NewRecorder()
			testHandler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("expected status 200, got %v", rr.Code)
			}
			if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
				t.Errorf("expected text/html, got %q", ct)
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

func TestProductsPageResponseHeaders(t *testing.T) {
	urls := []string{
		"/products",
		"/products?source=fptshop",
		"/products?category=Laptop",
		"/products?source=thegioididong&category=Laptop&sort=price,desc",
		"/products?min_price=0&max_price=0",
	}

	for _, url := range urls {
		t.Run(url, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			rr := httptest.NewRecorder()
			testHandler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("expected 200, got %v", rr.Code)
			}
			if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
				t.Errorf("expected text/html, got %q", ct)
			}
			if !strings.Contains(rr.Body.String(), "source-filter") {
				t.Errorf("expected source-filter panel in response for %s", url)
			}
		})
	}
}

func TestProductCardLinksCarryFilterParams(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet,
		"/products?source=fptshop&category=Laptop&subcategory=Acer", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "source=fptshop") {
		t.Error("expected product card links to contain source param")
	}
	if !strings.Contains(body, "category=Laptop") {
		t.Error("expected product card links to contain category param")
	}
	if !strings.Contains(body, "subcategory=Acer") {
		t.Error("expected product card links to contain subcategory param")
	}
}

func TestProductDetailBackURL(t *testing.T) {
	validID, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("failed to generate uuid: %v", err)
	}

	t.Run("back link includes filter params", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet,
			"/products/"+validID.String()+"?source=fptshop&category=Laptop&subcategory=Acer", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
		}
		body := rr.Body.String()
		if !strings.Contains(body, "/products?") {
			t.Error("expected back link to point to /products with params")
		}
		if !strings.Contains(body, "source=fptshop") {
			t.Error("expected back link to contain source param")
		}
		if !strings.Contains(body, "category=Laptop") {
			t.Error("expected back link to contain category param")
		}
	})

	t.Run("back link is plain /products when no filters", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet,
			"/products/"+validID.String(), nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
		}
		body := rr.Body.String()
		if !strings.Contains(body, `href="/products"`) {
			t.Error("expected back link to be plain /products")
		}
	})
}

func TestProductsFragmentCarriesFilterContext(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet,
		"/products-fragment?source=fptshop&category=Laptop&subcategory=Acer", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "source=fptshop") {
		t.Error("expected fragment card links to contain source param")
	}
	if !strings.Contains(body, "category=Laptop") {
		t.Error("expected fragment card links to contain category param")
	}
}
