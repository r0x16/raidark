package driver

import (
	"os"
	"testing"
)

func TestEnvProviderBasic(t *testing.T) {
	os.Setenv("STR", "val")
	defer os.Unsetenv("STR")
	p := NewEnvProvider()
	if v := p.GetString("STR", ""); v != "val" {
		t.Fatalf("expected val got %s", v)
	}
	if v := p.GetString("MISSING", "def"); v != "def" {
		t.Fatalf("expected def got %s", v)
	}
	os.Setenv("BOOL", "true")
	if !p.GetBool("BOOL", false) {
		t.Fatal("expected true")
	}
	os.Setenv("INT", "5")
	if p.GetInt("INT", 0) != 5 {
		t.Fatal("expected 5")
	}
	os.Setenv("FLOAT", "1.5")
	if p.GetFloat("FLOAT", 0) != 1.5 {
		t.Fatal("expected 1.5")
	}
	os.Setenv("SLICE", "a,b,c")
	s := p.GetSlice("SLICE", nil)
	if len(s) != 3 || s[1] != "b" {
		t.Fatalf("unexpected slice %v", s)
	}
	if !p.IsSet("STR") {
		t.Fatal("expected STR to be set")
	}
}

func TestEnvProviderMust(t *testing.T) {
	p := NewEnvProvider()
	os.Setenv("REQ", "x")
	defer os.Unsetenv("REQ")
	if p.MustGet("REQ") != "x" {
		t.Fatal("expected x")
	}
	os.Setenv("INTREQ", "3")
	if p.MustGetInt("INTREQ") != 3 {
		t.Fatal("expected 3")
	}
	os.Setenv("BOOLREQ", "true")
	if !p.MustGetBool("BOOLREQ") {
		t.Fatal("expected true")
	}
}

func TestEnvProviderMustPanics(t *testing.T) {
	defer func() { recover() }()
	p := NewEnvProvider()
	p.MustGet("NONE")
	t.Fatal("expected panic")
}
