package driver

import (
	"testing"
	"time"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/r0x16/Raidark/shared/auth/domain"
)

func TestArrayAuthProvider(t *testing.T) {
	p := NewArrayAuthProvider()
	if err := p.Initialize(); err != nil {
		t.Fatalf("init err %v", err)
	}
	if _, err := p.GetUser("admin"); err != nil {
		t.Fatalf("get user err %v", err)
	}
	u := &domain.User{User: casdoorsdk.User{Name: "new", CreatedTime: time.Now().Format(time.RFC3339), UpdatedTime: time.Now().Format(time.RFC3339)}}
	ok, err := p.AddUser(u)
	if !ok || err != nil {
		t.Fatalf("add user err %v", err)
	}
	ok, err = p.UpdateUser(u)
	if !ok || err != nil {
		t.Fatalf("update user err %v", err)
	}
	ok, err = p.DeleteUser(u)
	if !ok || err != nil {
		t.Fatalf("delete user err %v", err)
	}
	if err := p.HealthCheck(); err != nil {
		t.Fatalf("health err %v", err)
	}
}
