package auth

// ManagedAccount represents a managed account in the domain
type ManagedAccount struct {
	Application string
	Username    string
	Password    string
	SigninURL   string
}
