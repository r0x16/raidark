package domain

import "github.com/casdoor/casdoor-go-sdk/casdoorsdk"

// Permission represents a permission in the domain using composition with Casdoor SDK
type Permission struct {
	casdoorsdk.Permission // Embed Casdoor permission struct for direct access to all fields
}

// Business methods for domain logic
func (p *Permission) IsActive() bool {
	return p.IsEnabled && p.Name != ""
}

func (p *Permission) HasUser(username string) bool {
	for _, user := range p.Users {
		if user == username {
			return true
		}
	}
	return false
}

func (p *Permission) HasRole(roleName string) bool {
	for _, role := range p.Roles {
		if role == roleName {
			return true
		}
	}
	return false
}
