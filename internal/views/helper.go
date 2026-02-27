package views

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func extractFirstImage(raw string) string {
	var images []string
	if err := json.Unmarshal([]byte(raw), &images); err == nil && len(images) > 0 {
		return images[0]
	}
	return ""
}

func formatPrice(price string) string {
	f, err := strconv.ParseFloat(strings.TrimSpace(price), 64)
	if err != nil {
		return price
	}
	s := strconv.FormatInt(int64(f), 10)
	result := []byte{}
	for i, ch := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(ch))
	}
	return fmt.Sprintf("%s ₫", string(result))
}

func FormatMonthYear(t time.Time, lang string) string {
	if lang == "vi" {
		months := []string{
			"tháng 1", "tháng 2", "tháng 3", "tháng 4",
			"tháng 5", "tháng 6", "tháng 7", "tháng 8",
			"tháng 9", "tháng 10", "tháng 11", "tháng 12",
		}
		return months[t.Month()-1] + " năm " + strconv.Itoa(t.Year())
	}
	return t.Format("January 2006")
}

func FormatDateTime(t time.Time, lang string) string {
	if lang == "vi" {
		months := []string{
			"tháng 1", "tháng 2", "tháng 3", "tháng 4",
			"tháng 5", "tháng 6", "tháng 7", "tháng 8",
			"tháng 9", "tháng 10", "tháng 11", "tháng 12",
		}
		return strconv.Itoa(t.Day()) + " " + months[t.Month()-1] + ", " + strconv.Itoa(t.Year()) + " " + t.Format("15:04")
	}
	return t.Format("Jan 2, 2006 15:04")
}
