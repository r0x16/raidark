package domain

import "time"

// Token represents an OAuth token in the domain
type Token struct {
	AccessToken  string
	TokenType    string
	RefreshToken string
	Expiry       time.Time
}
