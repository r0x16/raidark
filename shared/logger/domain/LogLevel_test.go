package domain

import "testing"

func TestParseLogLevel(t *testing.T) {
	tests := map[string]LogLevel{
		"DEBUG":    Debug,
		"INFO":     Info,
		"WARNING":  Warning,
		"ERROR":    Error,
		"CRITICAL": Critical,
		"UNKNOWN":  Info,
	}
	for str, expected := range tests {
		if level := ParseLogLevel(str); level != expected {
			t.Errorf("%s expected %v got %v", str, expected, level)
		}
	}
}
