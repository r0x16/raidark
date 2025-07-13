package util

import (
	"testing"
)

func TestParseDate(t *testing.T) {
	d, err := ParseDate("01-01-2020")
	if err != nil || d == nil {
		t.Fatalf("expected date got %v %v", d, err)
	}
	if _, err := ParseDate("bad"); err == nil {
		t.Fatal("expected error")
	}
}

func TestParsePage(t *testing.T) {
	if ParsePage("2") != 2 || ParsePage("bad") != 1 || ParsePage("") != 1 {
		t.Fatal("unexpected parse page")
	}
}

func TestParsePageSize(t *testing.T) {
	if v, _ := ParsePageSize("50"); v != 50 {
		t.Fatal("expected 50")
	}
	if v, _ := ParsePageSize("bad"); v != 10 {
		t.Fatal("default 10")
	}
	if v, err := ParsePageSize("101"); v != 100 || err == nil {
		t.Fatal("expected capped 100 and error")
	}
}

func TestParseUintID(t *testing.T) {
	if v, _ := ParseUintID("5"); v != 5 {
		t.Fatal("expected 5")
	}
	if _, err := ParseUintID(""); err == nil {
		t.Fatal("expected error")
	}
}

func TestSanitizeString(t *testing.T) {
	s := SanitizeString(" <b> hi </b> ")
	if s != "&lt;b&gt; hi &lt;/b&gt;" {
		t.Fatalf("got %s", s)
	}
}
