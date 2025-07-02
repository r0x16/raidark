package domain

import "github.com/casdoor/casdoor-go-sdk/casdoorsdk"

// Role represents a user role in the domain using composition with Casdoor SDK
type Role struct {
	casdoorsdk.Role // Embed Casdoor role struct for direct access to all fields
}

// Business methods for domain logic
func (r *Role) IsActive() bool {
	return r.Name != "" && r.Owner != ""
}

func (r *Role) HasUser(username string) bool {
	for _, user := range r.Users {
		if user == username {
			return true
		}
	}
	return false
}
