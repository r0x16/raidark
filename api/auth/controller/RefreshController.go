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

// RefreshResponse represents the response structure for token refresh
type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

func RefreshAction(c echo.Context, bundle *drivers.ApplicationBundle) error {
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

	// Get session ID from cookie
	cookie, err := c.Cookie("raidark_session")
	if err != nil {
		bundle.Log.Warning("No session cookie found", map[string]any{
			"error": err.Error(),
		})
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "No valid session found",
		})
	}

	sessionID := cookie.Value
	if sessionID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid session",
		})
	}

	// Get user agent and IP address
	userAgent := c.Request().UserAgent()
	ipAddress := c.RealIP()

	// Initialize repository and auth service using dependency injection
	sessionRepo := repositories.NewGormSessionRepository(dbProvider.Connection)
	authService := service.NewAuthService(sessionRepo, bundle.Auth)

	// Attempt to refresh tokens
	session, token, err := authService.RefreshTokens(sessionID, userAgent, ipAddress)
	if err != nil {
		bundle.Log.Error("Failed to refresh tokens", map[string]any{
			"error":      err.Error(),
			"session_id": sessionID,
		})

		// Check if session was not found or expired
		if err.Error() == "session not found" || err.Error() == "refresh token expired" {
			// Clear the invalid cookie
			cookie := &http.Cookie{
				Name:     "raidark_session",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				MaxAge:   -1,
			}
			c.SetCookie(cookie)

			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Session expired or invalid",
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to refresh token",
		})
	}

	// Calculate expires in seconds
	expiresIn := int64(0)
	if !token.Expiry.IsZero() {
		expiresIn = int64(time.Until(token.Expiry).Seconds())
	}

	// Prepare successful response
	response := RefreshResponse{
		AccessToken: token.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
	}

	bundle.Log.Info("Token refresh successful", map[string]any{
		"session_id": sessionID,
		"user_id":    session.UserID,
		"username":   session.Username,
	})

	return c.JSON(http.StatusOK, response)
}
