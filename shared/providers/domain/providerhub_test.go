package domain

import "testing"

func TestRegisterAndGet(t *testing.T) {
	hub := &ProviderHub{}
	p := 42
	Register(hub, p)
	if !Exists[int](hub) {
		t.Fatalf("expected provider to exist")
	}
	v := Get[int](hub)
	if v != p {
		t.Fatalf("expected %d, got %d", p, v)
	}
}

func TestGetMissingPanics(t *testing.T) {
	hub := &ProviderHub{}
	defer func() { recover() }()
	_ = Get[int](hub)
}
