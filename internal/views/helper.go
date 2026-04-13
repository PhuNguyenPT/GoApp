package views

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
)

func extractFirstImage(raw []byte) string {
	var images []string
	if err := json.Unmarshal(raw, &images); err == nil && len(images) > 0 {
		return images[0]
	}
	return ""
}

func formatPrice(price string, currency string) string {
	f, err := strconv.ParseFloat(strings.TrimSpace(price), 64)
	if err != nil {
		return price
	}

	switch strings.ToUpper(strings.TrimSpace(currency)) {
	case "VND", "₫":
		s := strconv.FormatInt(int64(f), 10)
		result := []byte{}
		for i, ch := range s {
			if i > 0 && (len(s)-i)%3 == 0 {
				result = append(result, ',')
			}
			result = append(result, byte(ch))
		}
		return fmt.Sprintf("%s ₫", string(result))
	case "USD":
		return fmt.Sprintf("$%.2f", f)
	case "EUR":
		return fmt.Sprintf("€%.2f", f)
	case "GBP":
		return fmt.Sprintf("£%.2f", f)
	default:
		// fallback: use currency code as prefix with 2 decimal places
		return fmt.Sprintf("%s %.2f", strings.ToUpper(currency), f)
	}
}

func FormatMonthYear(t time.Time, lang string) string {
	if lang == "vi" {
		months := []string{
			"Tháng 1", "Tháng 2", "Tháng 3", "Tháng 4",
			"Tháng 5", "Tháng 6", "Tháng 7", "Tháng 8",
			"Tháng 9", "Tháng 10", "Tháng 11", "Tháng 12",
		}
		return months[t.Month()-1] + " năm " + strconv.Itoa(t.Year())
	}
	return t.Format("January 2006")
}

func FormatDateTime(t time.Time, lang string) string {
	if lang == "vi" {
		months := []string{
			"Tháng 1", "Tháng 2", "Tháng 3", "Tháng 4",
			"Tháng 5", "Tháng 6", "Tháng 7", "Tháng 8",
			"Tháng 9", "Tháng 10", "Tháng 11", "Tháng 12",
		}
		return strconv.Itoa(t.Day()) + " " + months[t.Month()-1] + ", " + strconv.Itoa(t.Year()) + " " + t.Format("15:04")
	}
	return t.Format("Jan 2, 2006 15:04")
}

func paginationPages(current, total int) []int {
	if total <= 1 {
		return nil
	}

	set := map[int]bool{}
	pages := []int{}

	add := func(p int) {
		if p >= 1 && p <= total && !set[p] {
			set[p] = true
		}
	}

	add(1)
	add(total)
	add(current)
	for i := -2; i <= 2; i++ {
		add(current + i)
	}

	// build sorted list
	sorted := []int{}
	for p := 1; p <= total; p++ {
		if set[p] {
			sorted = append(sorted, p)
		}
	}

	// insert 0 as ellipsis where gaps exist
	for i, p := range sorted {
		if i == 0 {
			pages = append(pages, p)
			continue
		}
		if p-pages[len(pages)-1] > 1 {
			pages = append(pages, 0) // ellipsis
		}
		pages = append(pages, p)
	}

	return pages
}

func extractAllImages(raw []byte) []string {
	var images []string
	if err := json.Unmarshal(raw, &images); err == nil {
		return images
	}
	return nil
}

func extractSpecsOrdered(raw json.RawMessage) [][2]string {
	var result [][2]string
	dec := json.NewDecoder(bytes.NewReader(raw))

	// Read opening {
	if t, err := dec.Token(); err != nil || t != json.Delim('{') {
		return nil
	}

	for dec.More() {
		// Read key
		keyToken, err := dec.Token()
		if err != nil {
			break
		}
		key, ok := keyToken.(string)
		if !ok {
			break
		}

		// Read value
		var value string
		if err := dec.Decode(&value); err != nil {
			break
		}

		result = append(result, [2]string{key, value})
	}

	return result
}

func filterItemClass(item, selected string) templ.CSSClasses {
	active := item == selected
	return templ.Classes(
		"flex items-center gap-2 text-sm rounded-lg px-2 py-1.5 w-full transition-colors",
		templ.KV("bg-blue-50", active),
		templ.KV("text-blue-700", active),
		templ.KV("font-medium", active),
		templ.KV("text-gray-600", !active),
		templ.KV("hover:bg-gray-50", !active),
	)
}

func priceItemClass(bMin, bMax, selMin, selMax float64) templ.CSSClasses {
	active := bMin == selMin && bMax == selMax
	return templ.Classes(
		"flex items-center gap-2 text-sm rounded-lg px-2 py-1.5 w-full transition-colors",
		templ.KV("bg-blue-50", active),
		templ.KV("text-blue-700", active),
		templ.KV("font-medium", active),
		templ.KV("text-gray-600", !active),
		templ.KV("hover:bg-gray-50", !active),
	)
}

func priceBucketsFromPercentiles(min, p25, p50, p75, max float64) [][2]float64 {
	points := []float64{min, p25, p50, p75, max}

	// deduplicate — if percentiles collapse (few products), skip duplicates
	var unique []float64
	for _, p := range points {
		if len(unique) == 0 || p > unique[len(unique)-1] {
			unique = append(unique, p)
		}
	}

	if len(unique) < 2 {
		return nil
	}

	var buckets [][2]float64
	for i := 0; i < len(unique)-1; i++ {
		buckets = append(buckets, [2]float64{unique[i], unique[i+1]})
	}
	return buckets
}
