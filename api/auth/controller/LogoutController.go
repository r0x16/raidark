package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/r0x16/Raidark/api/auth/drivers/repositories"
	"github.com/r0x16/Raidark/api/auth/service"
	"github.com/r0x16/Raidark/api/drivers"
	"github.com/r0x16/Raidark/shared/driver/db"
)

// LogoutResponse represents the response structure for logout
type LogoutResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func LogoutAction(c echo.Context, bundle *drivers.ApplicationBundle) error {
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
		// If no cookie exists, consider logout successful (already logged out)
		bundle.Log.Info("Logout attempt with no session cookie", map[string]any{
			"ip": c.RealIP(),
		})

		response := LogoutResponse{
			Message: "Already logged out",
			Success: true,
		}
		return c.JSON(http.StatusOK, response)
	}

	sessionID := cookie.Value
	if sessionID == "" {
		// Invalid session ID, clear cookie and return success
		clearCookie := &http.Cookie{
			Name:     "raidark_session",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1,
		}
		c.SetCookie(clearCookie)

		response := LogoutResponse{
			Message: "Invalid session cleared",
			Success: true,
		}
		return c.JSON(http.StatusOK, response)
	}

	// Initialize repository and auth service using dependency injection
	sessionRepo := repositories.NewGormSessionRepository(dbProvider.Connection)
	authService := service.NewAuthService(sessionRepo, bundle.Auth)

	// Invalidate session in database
	err = authService.InvalidateSession(sessionID)
	if err != nil {
		if err.Error() == "session not found" {
			// Session was already deleted, consider logout successful
			bundle.Log.Info("Logout attempt with non-existent session", map[string]any{
				"session_id": sessionID,
			})
		} else {
			// Log error but still proceed with cookie clearing
			bundle.Log.Error("Failed to invalidate session", map[string]any{
				"error":      err.Error(),
				"session_id": sessionID,
			})
		}
	}

	// Clear the session cookie regardless of database operation result
	clearCookie := &http.Cookie{
		Name:     "raidark_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	c.SetCookie(clearCookie)

	bundle.Log.Info("User logged out successfully", map[string]any{
		"session_id": sessionID,
		"ip":         c.RealIP(),
	})

	response := LogoutResponse{
		Message: "Logged out successfully",
		Success: true,
	}

	return c.JSON(http.StatusOK, response)
}
