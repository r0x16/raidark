package domain

// ExchangeRequest represents the request structure for token exchange
type ExchangeRequest struct {
	Code  string `query:"code"`
	State string `query:"state"`
}
