package util

import (
	"html"
	"strings"
)

func SanitizeString(str string) string {
	str = strings.TrimSpace(str)
	str = html.EscapeString(str)

	return str
}
