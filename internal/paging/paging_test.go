package paging

import "testing"

type ProductSortField string

const (
	ProductSortDefault   ProductSortField = ""
	ProductSortPrice     ProductSortField = "price"
	ProductSortCrawledAt ProductSortField = "crawled_at"
)

// --- Order (sort.go) ---

func TestOrder_AscDesc(t *testing.T) {
	asc := Order[ProductSortField]{Field: ProductSortPrice, Direction: Asc}
	desc := Order[ProductSortField]{Field: ProductSortPrice, Direction: Desc}
	if asc.Direction != Asc {
		t.Errorf("expected Asc, got %q", asc.Direction)
	}
	if desc.Direction != Desc {
		t.Errorf("expected Desc, got %q", desc.Direction)
	}
	if asc == desc {
		t.Error("asc and desc should not be equal")
	}
}

func TestOrder_ZeroValue(t *testing.T) {
	var o Order[ProductSortField]
	if o.Field != ProductSortDefault {
		t.Errorf("expected empty field, got %q", o.Field)
	}
	if o.Direction != "" {
		t.Errorf("expected empty dir, got %q", o.Direction)
	}
}

// --- NewPageable ---

func TestNewPageable(t *testing.T) {
	sort := Order[ProductSortField]{Field: ProductSortPrice, Direction: Asc}
	p := NewPageable(1, 20, sort)
	if p.Page != 1 {
		t.Errorf("expected page 1, got %d", p.Page)
	}
	if p.Size != 20 {
		t.Errorf("expected size 20, got %d", p.Size)
	}
	if p.Sort != sort {
		t.Errorf("expected sort to be set")
	}
}

// --- Offset ---

func TestPageable_Offset(t *testing.T) {
	tests := []struct {
		page       int
		size       int
		wantOffset int
	}{
		{1, 20, 0},
		{2, 20, 20},
		{3, 20, 40},
		{1, 50, 0},
		{3, 50, 100},
	}
	for _, tt := range tests {
		p := Pageable[ProductSortField]{Page: tt.page, Size: tt.size}
		if got := p.Offset(); got != tt.wantOffset {
			t.Errorf("page=%d size=%d: got offset %d, want %d", tt.page, tt.size, got, tt.wantOffset)
		}
	}
}

// --- Next ---

func TestPageable_Next(t *testing.T) {
	sort := Order[ProductSortField]{Field: ProductSortPrice, Direction: Asc}
	p := Pageable[ProductSortField]{Page: 2, Size: 20, Sort: sort}
	next := p.Next()
	if next.Page != 3 {
		t.Errorf("expected page 3, got %d", next.Page)
	}
	if next.Size != p.Size {
		t.Errorf("expected size %d, got %d", p.Size, next.Size)
	}
	if next.Sort != sort {
		t.Errorf("expected sort to be preserved")
	}
}

func TestPageable_Next_FromOne(t *testing.T) {
	p := Pageable[ProductSortField]{Page: 1, Size: 20}
	if p.Next().Page != 2 {
		t.Errorf("expected page 2, got %d", p.Next().Page)
	}
}

// --- PreviousOrFirst ---

func TestPageable_PreviousOrFirst_HasPrevious(t *testing.T) {
	p := Pageable[ProductSortField]{Page: 3, Size: 20}
	if p.PreviousOrFirst().Page != 2 {
		t.Errorf("expected page 2, got %d", p.PreviousOrFirst().Page)
	}
}

func TestPageable_PreviousOrFirst_AlreadyFirst(t *testing.T) {
	p := Pageable[ProductSortField]{Page: 1, Size: 20}
	if p.PreviousOrFirst().Page != 1 {
		t.Errorf("expected page 1, got %d", p.PreviousOrFirst().Page)
	}
}

// --- First ---

func TestPageable_First(t *testing.T) {
	sort := Order[ProductSortField]{Field: ProductSortPrice, Direction: Desc}
	p := Pageable[ProductSortField]{Page: 5, Size: 20, Sort: sort}
	first := p.First()
	if first.Page != 1 {
		t.Errorf("expected page 1, got %d", first.Page)
	}
	if first.Size != p.Size {
		t.Errorf("expected size %d, got %d", p.Size, first.Size)
	}
	if first.Sort != sort {
		t.Errorf("expected sort to be preserved")
	}
}

func TestPageable_First_AlreadyFirst(t *testing.T) {
	p := Pageable[ProductSortField]{Page: 1, Size: 20}
	if p.First().Page != 1 {
		t.Errorf("expected page 1, got %d", p.First().Page)
	}
}

// --- NewPage ---

func TestNewPage_TotalPages(t *testing.T) {
	tests := []struct {
		total     int64
		size      int
		wantPages int
	}{
		{0, 20, 1},
		{20, 20, 1},
		{21, 20, 2},
		{100, 20, 5},
		{101, 20, 6},
	}
	for _, tt := range tests {
		p := NewPage([]string(nil), tt.total, Pageable[ProductSortField]{Page: 1, Size: tt.size})
		if p.TotalPages != tt.wantPages {
			t.Errorf("total=%d size=%d: got %d pages, want %d", tt.total, tt.size, p.TotalPages, tt.wantPages)
		}
	}
}

func TestNewPage_TotalElements(t *testing.T) {
	p := NewPage([]string{"a", "b"}, int64(42), Pageable[ProductSortField]{Page: 1, Size: 20})
	if p.TotalElements != 42 {
		t.Errorf("expected TotalElements 42, got %d", p.TotalElements)
	}
}

func TestNewPage_PageablePreserved(t *testing.T) {
	pageable := Pageable[ProductSortField]{Page: 2, Size: 20, Sort: Order[ProductSortField]{Field: ProductSortPrice, Direction: Desc}}
	p := NewPage([]string(nil), int64(100), pageable)
	if p.Pageable != pageable {
		t.Error("expected pageable to be preserved")
	}
}

// --- HasNext / HasPrevious ---

func TestPage_HasNext(t *testing.T) {
	p := NewPage([]string(nil), int64(100), Pageable[ProductSortField]{Page: 4, Size: 20})
	if !p.HasNext() {
		t.Error("expected HasNext = true on page 4 of 5")
	}
	last := NewPage([]string(nil), int64(100), Pageable[ProductSortField]{Page: 5, Size: 20})
	if last.HasNext() {
		t.Error("expected HasNext = false on last page")
	}
}

func TestPage_HasNext_SinglePage(t *testing.T) {
	p := NewPage([]string(nil), int64(5), Pageable[ProductSortField]{Page: 1, Size: 20})
	if p.HasNext() {
		t.Error("expected HasNext = false when only 1 page")
	}
}

func TestPage_HasPrevious(t *testing.T) {
	p := NewPage([]string(nil), int64(100), Pageable[ProductSortField]{Page: 2, Size: 20})
	if !p.HasPrevious() {
		t.Error("expected HasPrevious = true on page 2")
	}
	first := NewPage([]string(nil), int64(100), Pageable[ProductSortField]{Page: 1, Size: 20})
	if first.HasPrevious() {
		t.Error("expected HasPrevious = false on first page")
	}
}

// --- IsFirst / IsLast ---

func TestPage_IsFirst(t *testing.T) {
	p := NewPage([]string(nil), int64(100), Pageable[ProductSortField]{Page: 1, Size: 20})
	if !p.IsFirst() {
		t.Error("expected IsFirst = true")
	}
	p2 := NewPage([]string(nil), int64(100), Pageable[ProductSortField]{Page: 2, Size: 20})
	if p2.IsFirst() {
		t.Error("expected IsFirst = false on page 2")
	}
}

func TestPage_IsLast(t *testing.T) {
	p := NewPage([]string(nil), int64(100), Pageable[ProductSortField]{Page: 5, Size: 20})
	if !p.IsLast() {
		t.Error("expected IsLast = true on last page")
	}
	p2 := NewPage([]string(nil), int64(100), Pageable[ProductSortField]{Page: 4, Size: 20})
	if p2.IsLast() {
		t.Error("expected IsLast = false on page 4 of 5")
	}
}

func TestPage_IsLast_SinglePage(t *testing.T) {
	p := NewPage([]string(nil), int64(5), Pageable[ProductSortField]{Page: 1, Size: 20})
	if !p.IsLast() {
		t.Error("expected IsLast = true when only 1 page")
	}
}

// --- IsEmpty ---

func TestPage_IsEmpty(t *testing.T) {
	p := NewPage([]string{}, int64(0), Pageable[ProductSortField]{Page: 1, Size: 20})
	if !p.IsEmpty() {
		t.Error("expected IsEmpty = true")
	}
	p2 := NewPage([]string{"a"}, int64(1), Pageable[ProductSortField]{Page: 1, Size: 20})
	if p2.IsEmpty() {
		t.Error("expected IsEmpty = false")
	}
}
