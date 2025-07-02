package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/r0x16/Raidark/api/drivers"
	"github.com/r0x16/Raidark/shared/auth/domain"
	"github.com/r0x16/Raidark/shared/auth/domain/model"
	"github.com/r0x16/Raidark/shared/auth/driver/repositories"
	"github.com/r0x16/Raidark/shared/auth/service"
	"github.com/r0x16/Raidark/shared/driver/db"
)

// RefreshController handles token refresh operations
type RefreshController struct {
	bundle *drivers.ApplicationBundle
}

// RefreshAction creates a RefreshController instance and delegates to the Refresh method
func RefreshAction(c echo.Context, bundle *drivers.ApplicationBundle) error {
	controller := &RefreshController{
		bundle: bundle,
	}
	return controller.Refresh(c)
}

// Refresh handles the token refresh process
// It validates the session, refreshes the access token, and updates the session cookie
func (rc *RefreshController) Refresh(c echo.Context) error {
	// Get database connection
	dbProvider, err := rc.getDatabaseProvider()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Internal server error",
		})
	}

	// Get session ID from cookie
	sessionID, err := rc.getSessionFromCookie(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "No valid session found",
		})
	}

	// Extract client information
	userAgent, ipAddress := rc.extractClientInfo(c)

	// Initialize auth service
	authService := rc.initializeAuthService(dbProvider)

	// Attempt to refresh tokens
	session, token, err := rc.refreshTokens(authService, sessionID, userAgent, ipAddress)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Token refresh failed",
		})
	}

	// Update session cookie expiry
	rc.updateSessionCookie(c, session)

	// Build and return response
	response := rc.buildResponse(token)

	rc.logSuccessfulRefresh(session)

	return c.JSON(http.StatusOK, response)
}

// getDatabaseProvider retrieves and validates the database provider from the bundle
func (rc *RefreshController) getDatabaseProvider() (*db.GormPostgresDatabaseProvider, error) {
	dbProvider, ok := rc.bundle.Database.(*db.GormPostgresDatabaseProvider)
	if !ok {
		rc.bundle.Log.Error("Failed to get database connection", map[string]any{
			"error": "invalid database provider type",
		})
		return nil, &echo.HTTPError{Code: http.StatusInternalServerError}
	}
	return dbProvider, nil
}

// getSessionFromCookie extracts and validates the session ID from the HTTP cookie
func (rc *RefreshController) getSessionFromCookie(c echo.Context) (string, error) {
	cookie, err := c.Cookie("app_session")
	if err != nil {
		rc.bundle.Log.Warning("No session cookie found", map[string]any{
			"error": err.Error(),
		})
		return "", err
	}

	sessionID := cookie.Value
	if sessionID == "" {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "invalid session")
	}

	return sessionID, nil
}

// extractClientInfo extracts user agent and IP address from the HTTP context
func (rc *RefreshController) extractClientInfo(c echo.Context) (string, string) {
	userAgent := c.Request().UserAgent()
	ipAddress := c.RealIP()
	return userAgent, ipAddress
}

// initializeAuthService creates and returns an instance of the refresh service
func (rc *RefreshController) initializeAuthService(dbProvider *db.GormPostgresDatabaseProvider) *service.AuthRefreshService {
	sessionRepo := repositories.NewGormSessionRepository(dbProvider.GetDataStore().Exec)
	return service.NewAuthRefreshService(sessionRepo, rc.bundle.Auth)
}

// refreshTokens attempts to refresh the access token using the session's refresh token
func (rc *RefreshController) refreshTokens(authService *service.AuthRefreshService, sessionID, userAgent, ipAddress string) (*model.AuthSession, *domain.Token, error) {
	session, token, err := authService.RefreshTokens(sessionID, userAgent, ipAddress)
	if err != nil {
		rc.bundle.Log.Warning("Failed to refresh tokens", map[string]any{
			"error":      err.Error(),
			"session_id": sessionID,
		})
		return nil, nil, err
	}
	return session, token, nil
}

// updateSessionCookie updates the session cookie with new expiry time
func (rc *RefreshController) updateSessionCookie(c echo.Context, session *model.AuthSession) {
	refreshCookie := &http.Cookie{
		Name:     "app_session",
		Value:    session.SessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
		Expires:  session.RefreshExpiry,
	}
	c.SetCookie(refreshCookie)
}

// buildResponse constructs the refresh response with the new access token
func (rc *RefreshController) buildResponse(token *domain.Token) domain.RefreshResponse {
	// Calculate token expiry
	expiresIn := int64(0)
	if !token.Expiry.IsZero() {
		expiresIn = int64(time.Until(token.Expiry).Seconds())
	}

	response := domain.RefreshResponse{
		AccessToken: token.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
	}

	return response
}

// logSuccessfulRefresh logs the successful token refresh operation
func (rc *RefreshController) logSuccessfulRefresh(session *model.AuthSession) {
	rc.bundle.Log.Info("Token refresh successful", map[string]any{
		"user_id":    session.UserID,
		"username":   session.Username,
		"session_id": session.SessionID,
	})
}
