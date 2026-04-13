package views

import (
	"testing"
	"time"
)

func TestFormatPrice(t *testing.T) {
	tests := []struct {
		price    string
		currency string
		expected string
	}{
		{"33590000.0", "VND", "33,590,000 ₫"},
		{"1000.0", "VND", "1,000 ₫"},
		{"0.0", "VND", "0 ₫"},
		{"invalid", "VND", "invalid"},
		{"  18690000.0  ", "VND", "18,690,000 ₫"},
		{"99.99", "USD", "$99.99"},
		{"49.99", "EUR", "€49.99"},
		{"29.99", "GBP", "£29.99"},
		{"100.00", "JPY", "JPY 100.00"},
		{"100.00", "", " 100.00"},
	}
	for _, tt := range tests {
		got := formatPrice(tt.price, tt.currency)
		if got != tt.expected {
			t.Errorf("formatPrice(%q, %q) = %q, want %q", tt.price, tt.currency, got, tt.expected)
		}
	}
}

func TestFormatMonthYear(t *testing.T) {
	ts := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	if got := FormatMonthYear(ts, "en"); got != "March 2024" {
		t.Errorf("expected 'March 2024', got %q", got)
	}
	if got := FormatMonthYear(ts, "vi"); got != "Tháng 3 năm 2024" {
		t.Errorf("expected 'Tháng 3 năm 2024', got %q", got)
	}
}

func TestFormatDateTime(t *testing.T) {
	ts := time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC)
	if got := FormatDateTime(ts, "en"); got != "Mar 15, 2024 14:30" {
		t.Errorf("expected 'Mar 15, 2024 14:30', got %q", got)
	}
	if got := FormatDateTime(ts, "vi"); got != "15 Tháng 3, 2024 14:30" {
		t.Errorf("expected '15 Tháng 3, 2024 14:30', got %q", got)
	}
}

func TestPaginationPages(t *testing.T) {
	tests := []struct {
		current  int
		total    int
		expected []int
	}{
		{1, 1, nil},
		{1, 5, []int{1, 2, 3, 0, 5}},
		{1, 20, []int{1, 2, 3, 0, 20}},
		{8, 20, []int{1, 0, 6, 7, 8, 9, 10, 0, 20}},
		{19, 20, []int{1, 0, 17, 18, 19, 20}},
		{20, 20, []int{1, 0, 18, 19, 20}},
		{3, 5, []int{1, 2, 3, 4, 5}},
	}
	for _, tt := range tests {
		got := paginationPages(tt.current, tt.total)
		if !equalSlices(got, tt.expected) {
			t.Errorf("paginationPages(%d, %d) = %v, want %v", tt.current, tt.total, got, tt.expected)
		}
	}
}

func equalSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
