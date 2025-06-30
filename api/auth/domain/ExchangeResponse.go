package domain

// ExchangeResponse represents the response structure for token exchange
// NOTE: RefreshToken is intentionally NOT included for security reasons.
// Refresh tokens must never be exposed to the frontend and are stored securely in the backend session.
type ExchangeResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	User        struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Name     string `json:"name"`
		Email    string `json:"email"`
	} `json:"user"`
}
