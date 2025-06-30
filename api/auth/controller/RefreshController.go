package controller

import (
	"net/http"

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
	_, _, err = authService.RefreshTokens(sessionID, userAgent, ipAddress)
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

		// For the not implemented error, return appropriate status
		if err.Error() == "refresh token functionality not yet implemented - requires custom HTTP client for Casdoor" {
			return c.JSON(http.StatusNotImplemented, map[string]string{
				"error": "Refresh token functionality not yet implemented",
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to refresh token",
		})
	}

	// This part would be executed once refresh functionality is implemented
	// For now, it's unreachable due to the error above
	response := RefreshResponse{
		AccessToken: "new_access_token_would_be_here",
		TokenType:   "Bearer",
		ExpiresIn:   3600, // 1 hour
	}

	bundle.Log.Info("Token refresh successful", map[string]any{
		"session_id": sessionID,
	})

	return c.JSON(http.StatusOK, response)
}
