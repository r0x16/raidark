package domain

import "testing"

func TestRegisterAndGetProvider(t *testing.T) {
	hub := &ProviderHub{}
	type sample struct{ Value string }
	expected := sample{"test"}
	Register(hub, expected)
	got := Get[sample](hub)
	if got != expected {
		t.Fatalf("expected %v got %v", expected, got)
	}
}

func TestProviderExists(t *testing.T) {
	hub := &ProviderHub{}
	type sample struct{}
	if Exists[sample](hub) {
		t.Fatal("provider should not exist")
	}
	Register(hub, sample{})
	if !Exists[sample](hub) {
		t.Fatal("provider should exist")
	}
}

func TestGetMissingProviderPanics(t *testing.T) {
	defer func() { recover() }()
	hub := &ProviderHub{}
	type sample struct{}
	Get[sample](hub)
	t.Fatal("expected panic")
}
