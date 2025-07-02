package domain

// ExchangeRequest represents the request structure for token exchange
type ExchangeRequest struct {
	Code  string `json:"code" form:"code"`
	State string `json:"state" form:"state"`
}
