package utils

import (
	"fmt"
	"net/http"
	"strings"
)

func IsHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func Slugify(title string) string {
	s := strings.ToLower(title)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, s)
	return s
}

func BuildCanonicalURL(r *http.Request, path string) string {
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s%s", scheme, r.Host, path)
}
