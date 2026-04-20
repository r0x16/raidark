package domain

// Claims represents JWT claims in the domain
type Claims struct {
	Username     string
	Name         string
	Email        string
	Organization string
	Roles        []string
	Type         string
	Issuer       string
	Subject      string
	Audience     string
	ExpiresAt    int64
	IssuedAt     int64
	NotBefore    int64
}
