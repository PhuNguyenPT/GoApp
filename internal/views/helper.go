package views

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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
