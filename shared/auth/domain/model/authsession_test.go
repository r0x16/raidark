package model

import (
	"testing"
	"time"
)

func TestExpiryChecks(t *testing.T) {
	session := &AuthSession{ExpiresAt: time.Now().Add(-time.Hour), RefreshExpiry: time.Now().Add(-time.Hour)}
	if !session.IsExpired() {
		t.Fatal("expected expired")
	}
	if !session.IsRefreshExpired() {
		t.Fatal("expected refresh expired")
	}
}
