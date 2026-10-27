package domain

import (
	"net/mail"
	"net/url"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// CleanText trims s, drops control characters (newlines become spaces) and cuts it to max runes.
func CleanText(s string, max int) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '\n' || r == '\r' || r == '\t':
			b.WriteRune(' ')
		case unicode.IsControl(r):
			// dropped
		default:
			b.WriteRune(r)
		}
	}
	out := strings.TrimSpace(b.String())
	if utf8.RuneCountInString(out) > max {
		out = string([]rune(out)[:max])
	}
	return out
}

// CleanMultiline is CleanText that keeps line breaks (for descriptions).
func CleanMultiline(s string, max int) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	for _, r := range s {
		if r == '\n' || r == '\t' || !unicode.IsControl(r) {
			b.WriteRune(r)
		}
	}
	out := strings.TrimSpace(b.String())
	if utf8.RuneCountInString(out) > max {
		out = string([]rune(out)[:max])
	}
	return out
}

// ValidEmail accepts a plain address ("a@b.co"), not "Name <a@b.co>".
func ValidEmail(s string) bool {
	if s == "" || len(s) > 254 || strings.ContainsAny(s, " <>,;") {
		return false
	}
	a, err := mail.ParseAddress(s)
	return err == nil && a.Address == s && strings.Contains(s[strings.LastIndex(s, "@"):], ".")
}

// ValidPhone accepts digits with an optional leading +, spaces, dashes and parentheses (6-20 chars).
func ValidPhone(s string) bool {
	if len(s) < 6 || len(s) > 20 {
		return false
	}
	digits := 0
	for i, r := range s {
		switch {
		case r >= '0' && r <= '9':
			digits++
		case r == '+' && i == 0, r == ' ', r == '-', r == '(', r == ')':
		default:
			return false
		}
	}
	return digits >= 6
}

// ValidDay reports whether s is a calendar day in YYYY-MM-DD form.
func ValidDay(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

// ValidCurrency reports whether s is a three-letter upper-case code such as COP.
func ValidCurrency(s string) bool {
	if len(s) != 3 {
		return false
	}
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

// ValidHTTPURL reports whether s is an absolute http(s) URL.
func ValidHTTPURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" && len(s) <= 500
}

// CleanList trims entries, drops blanks and duplicates (case-insensitive) and keeps at most n.
func CleanList(in []string, n, maxLen int) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, s := range in {
		s = CleanText(s, maxLen)
		k := Normalize(s)
		if s == "" || seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, s)
		if len(out) == n {
			break
		}
	}
	return out
}
