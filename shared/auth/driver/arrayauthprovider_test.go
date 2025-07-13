package driver

import (
	"testing"

	domain "github.com/r0x16/Raidark/shared/auth/domain"
)

func TestArrayAuthProviderUserLifecycle(t *testing.T) {
	a := NewArrayAuthProvider()
	if err := a.Initialize(); err != nil {
		t.Fatalf("init err: %v", err)
	}
	u, err := a.GetUser("admin")
	if err != nil || u == nil {
		t.Fatalf("expected user")
	}
	newUser := domain.User{User: u.User}
	newUser.Name = "newuser"
	newUser.Id = "new-id"
	added, err := a.AddUser(&newUser)
	if err != nil || !added {
		t.Fatalf("add failed")
	}
	got, err := a.GetUser("newuser")
	if err != nil || got == nil {
		t.Fatalf("get failed")
	}
	ok, err := a.DeleteUser(got)
	if !ok || err != nil {
		t.Fatalf("delete failed")
	}
}
