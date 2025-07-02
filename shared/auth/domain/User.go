package domain

import "github.com/casdoor/casdoor-go-sdk/casdoorsdk"

// User represents a user in the domain using composition with Casdoor SDK
type User struct {
	casdoorsdk.User // Embed Casdoor user struct for direct access to all fields
}

// Business methods for domain logic
func (u *User) IsActive() bool {
	return !u.IsDeleted && !u.IsForbidden
}

func (u *User) IsVerified() bool {
	return u.EmailVerified
}

func (u *User) HasAdminRights() bool {
	return u.IsAdmin
}

func (u *User) GetFullName() string {
	if u.FirstName != "" && u.LastName != "" {
		return u.FirstName + " " + u.LastName
	}
	return u.DisplayName
}
