package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/r0x16/Raidark/shared/auth/domain"
	"github.com/r0x16/Raidark/shared/auth/driver/repositories"
	"github.com/r0x16/Raidark/shared/auth/service"
	domdatastore "github.com/r0x16/Raidark/shared/datastore/domain"
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
)

// LogoutController handles user logout operations
type LogoutController struct {
	Datastore domdatastore.DatabaseProvider
	Auth      domain.AuthProvider
	Log       domlogger.LogProvider
}

// LogoutAction creates a LogoutController instance and delegates to the Logout method
func LogoutAction(c echo.Context, hub *domprovider.ProviderHub) error {
	controller := &LogoutController{
		Datastore: domprovider.Get[domdatastore.DatabaseProvider](hub),
		Auth:      domprovider.Get[domain.AuthProvider](hub),
		Log:       domprovider.Get[domlogger.LogProvider](hub),
	}
	return controller.Logout(c)
}

// Logout handles the user logout process
// It invalidates the session in the database and clears the session cookie
func (lc *LogoutController) Logout(c echo.Context) error {

	// Get session ID from cookie
	sessionID, err := lc.getSessionFromCookie(c)
	if err != nil {
		// Handle cases where no session exists
		return lc.handleNoSession(c, err)
	}

	// Initialize auth service
	authService := lc.initializeAuthService(lc.Datastore)

	// Invalidate session in database
	err = lc.invalidateSession(authService, sessionID)
	if err != nil {
		lc.handleSessionInvalidationError(sessionID, err)
	}

	// Clear session cookie
	lc.clearSessionCookie(c)

	// Log successful logout
	lc.logSuccessfulLogout(sessionID, c.RealIP())

	// Return success response
	response := domain.LogoutResponse{
		Message: "Logged out successfully",
		Success: true,
	}

	return c.JSON(http.StatusOK, response)
}

// getSessionFromCookie extracts the session ID from the HTTP cookie
func (lc *LogoutController) getSessionFromCookie(c echo.Context) (string, error) {
	cookie, err := c.Cookie("app_session")
	if err != nil {
		return "", err
	}

	sessionID := cookie.Value
	if sessionID == "" {
		return "", echo.NewHTTPError(http.StatusBadRequest, "empty session ID")
	}

	return sessionID, nil
}

// handleNoSession handles cases where no valid session cookie exists
func (lc *LogoutController) handleNoSession(c echo.Context, err error) error {
	if err.Error() == "http: named cookie not present" {
		// If no cookie exists, consider logout successful (already logged out)
		lc.Log.Info("Logout attempt with no session cookie", map[string]any{
			"ip": c.RealIP(),
		})

		response := domain.LogoutResponse{
			Message: "Already logged out",
			Success: true,
		}
		return c.JSON(http.StatusOK, response)
	}

	// Invalid session ID, clear cookie and return success
	lc.clearSessionCookie(c)

	response := domain.LogoutResponse{
		Message: "Invalid session cleared",
		Success: true,
	}
	return c.JSON(http.StatusOK, response)
}

// initializeAuthService creates and returns an instance of the logout service
func (lc *LogoutController) initializeAuthService(dbProvider domdatastore.DatabaseProvider) *service.AuthLogoutService {
	sessionRepo := repositories.NewGormSessionRepository(dbProvider.GetDataStore().Exec)
	return service.NewAuthLogoutService(sessionRepo, lc.Auth)
}

// invalidateSession attempts to invalidate the session in the database
func (lc *LogoutController) invalidateSession(authService *service.AuthLogoutService, sessionID string) error {
	return authService.InvalidateSession(sessionID)
}

// handleSessionInvalidationError handles errors that occur during session invalidation
func (lc *LogoutController) handleSessionInvalidationError(sessionID string, err error) {
	if err.Error() == "session not found" {
		// Session was already deleted, consider logout successful
		lc.Log.Info("Logout attempt with non-existent session", map[string]any{
			"session_id": sessionID,
		})
	} else {
		// Log error but still proceed with cookie clearing
		lc.Log.Error("Failed to invalidate session", map[string]any{
			"error":      err.Error(),
			"session_id": sessionID,
		})
	}
}

// clearSessionCookie clears the session cookie from the HTTP response
func (lc *LogoutController) clearSessionCookie(c echo.Context) {
	clearCookie := &http.Cookie{
		Name:     "app_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	c.SetCookie(clearCookie)
}

// logSuccessfulLogout logs the successful logout operation
func (lc *LogoutController) logSuccessfulLogout(sessionID, ipAddress string) {
	lc.Log.Info("User logged out successfully", map[string]any{
		"session_id": sessionID,
		"ip":         ipAddress,
	})
}
