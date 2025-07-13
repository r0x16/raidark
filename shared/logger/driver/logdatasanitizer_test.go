package logger

import "testing"

func TestSanitizeValue(t *testing.T) {
	s := NewLogDataSanitizer()
	if v := s.SanitizeValue("text"); v != "text" {
		t.Fatalf("expected same string")
	}
	complex := struct{ A int }{A: 1}
	res := s.SanitizeValue(complex)
	if res == nil {
		t.Fatalf("expected sanitized string")
	}
}

func TestParseDataForSlog(t *testing.T) {
	s := NewLogDataSanitizer()
	attrs := s.ParseDataForSlog(map[string]any{"key": 1})
	if len(attrs) != 1 {
		t.Fatalf("expected 1 attr")
	}
}
