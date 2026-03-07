package server

import (
	"GoApp/internal/paging"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
)

type testSortField string

const (
	testSortDefault testSortField = ""
	testSortA       testSortField = "field_a"
	testSortB       testSortField = "field_b"
)

func newContext(params map[string]string) *gin.Context {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	req := httptest.NewRequest(http.MethodGet, "/?"+values.Encode(), nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return c
}

// --- page ---

func TestResolvePageable_DefaultPage(t *testing.T) {
	c := newContext(map[string]string{})
	p := resolvePageable[testSortField](c, 20, 100)
	if p.Page != 1 {
		t.Errorf("expected page 1, got %d", p.Page)
	}
}

func TestResolvePageable_ValidPage(t *testing.T) {
	c := newContext(map[string]string{"page": "3"})
	p := resolvePageable[testSortField](c, 20, 100)
	if p.Page != 3 {
		t.Errorf("expected page 3, got %d", p.Page)
	}
}

func TestResolvePageable_NegativePage_FallsBackToOne(t *testing.T) {
	c := newContext(map[string]string{"page": "-1"})
	p := resolvePageable[testSortField](c, 20, 100)
	if p.Page != 1 {
		t.Errorf("expected page 1, got %d", p.Page)
	}
}

func TestResolvePageable_InvalidPage_FallsBackToOne(t *testing.T) {
	c := newContext(map[string]string{"page": "abc"})
	p := resolvePageable[testSortField](c, 20, 100)
	if p.Page != 1 {
		t.Errorf("expected page 1, got %d", p.Page)
	}
}

// --- size ---

func TestResolvePageable_DefaultSize(t *testing.T) {
	c := newContext(map[string]string{})
	p := resolvePageable[testSortField](c, 20, 100)
	if p.Size != 20 {
		t.Errorf("expected default size 20, got %d", p.Size)
	}
}

func TestResolvePageable_ValidSize(t *testing.T) {
	c := newContext(map[string]string{"size": "50"})
	p := resolvePageable[testSortField](c, 20, 100)
	if p.Size != 50 {
		t.Errorf("expected size 50, got %d", p.Size)
	}
}

func TestResolvePageable_SizeExceedsMax_FallsBackToDefault(t *testing.T) {
	c := newContext(map[string]string{"size": "999"})
	p := resolvePageable[testSortField](c, 20, 100)
	if p.Size != 20 {
		t.Errorf("expected default size 20, got %d", p.Size)
	}
}

func TestResolvePageable_InvalidSize_FallsBackToDefault(t *testing.T) {
	c := newContext(map[string]string{"size": "abc"})
	p := resolvePageable[testSortField](c, 20, 100)
	if p.Size != 20 {
		t.Errorf("expected default size 20, got %d", p.Size)
	}
}

func TestResolvePageable_ZeroSize_FallsBackToDefault(t *testing.T) {
	c := newContext(map[string]string{"size": "0"})
	p := resolvePageable[testSortField](c, 20, 100)
	if p.Size != 20 {
		t.Errorf("expected default size 20, got %d", p.Size)
	}
}

// --- sort ---

func TestResolvePageable_ValidSort(t *testing.T) {
	tests := []struct {
		sort      string
		wantField testSortField
		wantDir   paging.Direction
	}{
		{"field_a,asc", testSortA, paging.Asc},
		{"field_a,desc", testSortA, paging.Desc},
		{"field_b,asc", testSortB, paging.Asc},
		{"field_b,desc", testSortB, paging.Desc},
	}
	for _, tt := range tests {
		c := newContext(map[string]string{"sort": tt.sort})
		p := resolvePageable[testSortField](c, 20, 100)
		if p.Sort.Field != tt.wantField {
			t.Errorf("sort=%q: got field %q, want %q", tt.sort, p.Sort.Field, tt.wantField)
		}
		if p.Sort.Direction != tt.wantDir {
			t.Errorf("sort=%q: got dir %q, want %q", tt.sort, p.Sort.Direction, tt.wantDir)
		}
	}
}

func TestResolvePageable_InvalidSortDir_FallsBackToAsc(t *testing.T) {
	c := newContext(map[string]string{"sort": "field_a,sideways"})
	p := resolvePageable[testSortField](c, 20, 100)
	if p.Sort.Direction != paging.Asc {
		t.Errorf("expected asc fallback, got %q", p.Sort.Direction)
	}
}

func TestResolvePageable_EmptySort_FallsBackToAsc(t *testing.T) {
	c := newContext(map[string]string{})
	p := resolvePageable[testSortField](c, 20, 100)
	if p.Sort.Direction != paging.Asc {
		t.Errorf("expected asc fallback, got %q", p.Sort.Direction)
	}
}

func TestResolvePageable_SQLInjection(t *testing.T) {
	malicious := []string{
		"field_a; DROP TABLE products--,asc",
		"field_a::numeric,asc",
		"' OR '1'='1,asc",
	}
	for _, input := range malicious {
		c := newContext(map[string]string{"sort": input})
		p := resolvePageable[testSortField](c, 20, 100)
		if p.Sort.Direction != paging.Asc && p.Sort.Direction != paging.Desc {
			t.Errorf("input %q: got unsafe direction %q", input, p.Sort.Direction)
		}
	}
}
