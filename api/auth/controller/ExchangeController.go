package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/r0x16/Raidark/api/auth/drivers/repositories"
	"github.com/r0x16/Raidark/api/auth/service"
	"github.com/r0x16/Raidark/api/drivers"
	"github.com/r0x16/Raidark/shared/driver/db"
)

// ExchangeRequest represents the request structure for token exchange
type ExchangeRequest struct {
	Code  string `json:"code" form:"code"`
	State string `json:"state" form:"state"`
}

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

func ExchangeAction(c echo.Context, bundle *drivers.ApplicationBundle) error {
	// Get database connection
	dbProvider, ok := bundle.Database.(*db.GormPostgresDatabaseProvider)
	if !ok {
		bundle.Log.Error("Failed to get database connection", map[string]any{
			"error": "invalid database provider type",
		})
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Internal server error",
		})
	}

	// Parse request
	var req ExchangeRequest
	if err := c.Bind(&req); err != nil {
		bundle.Log.Warning("Invalid exchange request", map[string]any{
			"error": err.Error(),
		})
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request parameters",
		})
	}

	// Validate required parameters
	if req.Code == "" || req.State == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing required parameters: code and state",
		})
	}

	// Get user agent and IP address
	userAgent := c.Request().UserAgent()
	ipAddress := c.RealIP()

	// Initialize repository and auth service using dependency injection
	sessionRepo := repositories.NewGormSessionRepository(dbProvider.Connection)
	authService := service.NewAuthExchangeService(sessionRepo, bundle.Auth)

	// Exchange code for tokens and create session
	session, token, claims, err := authService.ExchangeCodeForTokens(req.Code, req.State, userAgent, ipAddress)
	if err != nil {
		bundle.Log.Error("Failed to exchange code for tokens", map[string]any{
			"error": err.Error(),
			"code":  req.Code,
			"state": req.State,
		})
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Authentication failed",
		})
	}

	// Set secure cookie with session ID
	cookie := &http.Cookie{
		Name:     "raidark_session",
		Value:    session.SessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
		Expires:  session.RefreshExpiry,
	}
	c.SetCookie(cookie)

	// Prepare response
	expiresIn := int64(0)
	if !token.Expiry.IsZero() {
		expiresIn = int64(time.Until(token.Expiry).Seconds())
	}

	response := ExchangeResponse{
		AccessToken: token.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
	}

	// Set user information from claims
	response.User.ID = claims.Subject
	response.User.Username = claims.Username
	response.User.Name = claims.Name
	response.User.Email = claims.Email

	bundle.Log.Info("Token exchange successful", map[string]any{
		"user_id":    claims.Subject,
		"username":   claims.Username,
		"session_id": session.SessionID,
	})

	return c.JSON(http.StatusOK, response)
}
