package auth

// Role represents a user role in the domain
type Role struct {
	Owner       string
	Name        string
	CreatedTime string
	DisplayName string
	Description string
	Users       []string
	Groups      []string
	Domains     []string
}
