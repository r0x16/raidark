package driver

import (
	"os"
	"testing"
)

func TestGetStringWithDefault(t *testing.T) {
	os.Setenv("TEST_STRING", "value")
	defer os.Unsetenv("TEST_STRING")
	e := NewEnvProvider()
	if v := e.GetString("TEST_STRING", "default"); v != "value" {
		t.Fatalf("expected value, got %s", v)
	}
	if v := e.GetString("MISSING", "default"); v != "default" {
		t.Fatalf("expected default, got %s", v)
	}
}

func TestGetBool(t *testing.T) {
	os.Setenv("TEST_BOOL", "true")
	defer os.Unsetenv("TEST_BOOL")
	e := NewEnvProvider()
	if !e.GetBool("TEST_BOOL", false) {
		t.Fatal("expected true")
	}
	if e.GetBool("MISSING", true) != true {
		t.Fatal("expected default true")
	}
}

func TestMustGetPanics(t *testing.T) {
	e := NewEnvProvider()
	defer func() { recover() }()
	_ = e.MustGet("MISSING")
}
