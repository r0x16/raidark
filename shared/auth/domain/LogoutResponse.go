package domain

// LogoutResponse represents the response structure for logout
type LogoutResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}
