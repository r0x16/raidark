package logger

import (
	"testing"
)

type complex struct {
	A string
	B int
}

func TestSanitizeValue(t *testing.T) {
	s := NewLogDataSanitizer()
	if v := s.SanitizeValue("test"); v != "test" {
		t.Fatalf("expected test got %v", v)
	}
	v := s.SanitizeValue(complex{"a", 1}).(string)
	if len(v) == 0 {
		t.Fatal("expected string")
	}
}

func TestSanitizeDataAndParse(t *testing.T) {
	s := NewLogDataSanitizer()
	data := map[string]any{"x": complex{"a", 1}}
	sanitized := s.SanitizeData(data)
	if _, ok := sanitized["x"].(string); !ok {
		t.Fatal("expected sanitized string")
	}
	attrs := s.ParseDataForSlog(data)
	if len(attrs) == 0 {
		t.Fatal("expected attrs")
	}
}
