package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/r0x16/Raidark/shared/auth/domain"
	"github.com/r0x16/Raidark/shared/auth/domain/model"
	"github.com/r0x16/Raidark/shared/auth/driver/repositories"
	"github.com/r0x16/Raidark/shared/auth/service"
	domdatastore "github.com/r0x16/Raidark/shared/datastore/domain"
	domevents "github.com/r0x16/Raidark/shared/events/domain"
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
)

// ExchangeController handles OAuth2 authorization code exchange operations
type ExchangeController struct {
	Datastore domdatastore.DatabaseProvider
	Auth      domain.AuthProvider
	Log       domlogger.LogProvider
	Events    domevents.DomainEventsProvider
}

// ExchangeAction creates an ExchangeController instance and delegates to the Exchange method
func ExchangeAction(c echo.Context, hub *domprovider.ProviderHub) error {
	// Validate that AuthProvider exists in the hub
	if !domprovider.Exists[domain.AuthProvider](hub) {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Authentication provider not configured",
		})
	}

	// Validate that required providers exist
	if !domprovider.Exists[domdatastore.DatabaseProvider](hub) {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database provider not configured",
		})
	}

	if !domprovider.Exists[domlogger.LogProvider](hub) {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Logger provider not configured",
		})
	}

	var events domevents.DomainEventsProvider = nil
	if domprovider.Exists[domevents.DomainEventsProvider](hub) {
		events = domprovider.Get[domevents.DomainEventsProvider](hub)
	}

	controller := &ExchangeController{
		Datastore: domprovider.Get[domdatastore.DatabaseProvider](hub),
		Auth:      domprovider.Get[domain.AuthProvider](hub),
		Log:       domprovider.Get[domlogger.LogProvider](hub),
		Events:    events,
	}
	return controller.Exchange(c)
}

// Exchange handles the OAuth2 authorization code exchange process
// It validates the request, exchanges the code for tokens, creates a session,
// and returns the access token with user information
func (ec *ExchangeController) Exchange(c echo.Context) error {
	// Parse and validate request
	req, err := ec.parseRequest(c)
	if err != nil {
		return err
	}

	// Additional validation - req should not be nil after parseRequest
	if req == nil {
		ec.Log.Error("Request is nil after parseRequest", nil)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request",
		})
	}

	// Extract client information
	userAgent, ipAddress := ec.extractClientInfo(c)

	// Validate client information
	if userAgent == "" {
		ec.Log.Warning("User agent is empty", nil)
		userAgent = "unknown"
	}

	if ipAddress == "" {
		ec.Log.Warning("IP address is empty", nil)
		ipAddress = "unknown"
	}

	// Initialize services
	authService := ec.initializeAuthService(ec.Datastore)
	if authService == nil {
		ec.Log.Error("Failed to initialize authentication service", nil)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Authentication service unavailable",
		})
	}

	// Exchange code for tokens and create session
	session, token, claims, err := ec.exchangeCodeForTokens(authService, req, userAgent, ipAddress)
	if err != nil {
		ec.Log.Error("exchangeCodeForTokens failed", map[string]any{
			"error": err.Error(),
		})
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Authentication failed",
		})
	}

	// Validate returned values
	if session == nil || token == nil || claims == nil {
		ec.Log.Error("exchangeCodeForTokens returned nil values", map[string]any{
			"session_nil": session == nil,
			"token_nil":   token == nil,
			"claims_nil":  claims == nil,
		})
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Authentication processing failed",
		})
	}

	// Set session cookie
	ec.setSessionCookie(c, session)

	// Build and return response
	response := ec.buildResponse(token, claims)

	ec.logSuccessfulExchange(claims, session)

	return c.JSON(http.StatusOK, response)
}

// parseRequest parses and validates the exchange request from the HTTP context
func (ec *ExchangeController) parseRequest(c echo.Context) (*domain.ExchangeRequest, error) {
	var req domain.ExchangeRequest
	if err := c.Bind(&req); err != nil {
		ec.Log.Warning("Invalid exchange request", map[string]any{
			"error": err.Error(),
		})
		return nil, c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request parameters",
		})
	}

	// Validate required parameters
	if req.Code == "" || req.State == "" {
		return nil, c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing required parameters: code and state",
		})
	}

	return &req, nil
}

// extractClientInfo extracts user agent and IP address from the HTTP context
func (ec *ExchangeController) extractClientInfo(c echo.Context) (string, string) {
	userAgent := c.Request().UserAgent()
	ipAddress := c.RealIP()
	return userAgent, ipAddress
}

// initializeAuthService creates and returns an instance of the authentication service
func (ec *ExchangeController) initializeAuthService(dbProvider domdatastore.DatabaseProvider) *service.AuthExchangeService {
	// Additional safety check - this should not happen if ExchangeAction validation works correctly
	if ec.Auth == nil {
		ec.Log.Error("AuthProvider is nil in ExchangeController", nil)
		return nil
	}

	// Validate database provider
	if dbProvider == nil {
		ec.Log.Error("DatabaseProvider is nil in initializeAuthService", nil)
		return nil
	}

	// Validate database provider methods
	if dbProvider.GetDataStore() == nil {
		ec.Log.Error("DatabaseProvider.GetDataStore() returned nil", nil)
		return nil
	}

	if dbProvider.GetDataStore().Exec == nil {
		ec.Log.Error("DatabaseProvider.GetDataStore().Exec is nil", nil)
		return nil
	}

	sessionRepo := repositories.NewGormSessionRepository(dbProvider.GetDataStore().Exec)
	if sessionRepo == nil {
		ec.Log.Error("Failed to create session repository", nil)
		return nil
	}

	return service.NewAuthExchangeService(sessionRepo, ec.Auth, ec.Events)
}

// exchangeCodeForTokens performs the OAuth2 code exchange and returns session, token, and claims
func (ec *ExchangeController) exchangeCodeForTokens(authService *service.AuthExchangeService, req *domain.ExchangeRequest, userAgent, ipAddress string) (*model.AuthSession, *domain.Token, *domain.Claims, error) {
	// Validate all parameters to prevent nil pointer dereference
	if authService == nil {
		ec.Log.Error("AuthService is nil in exchangeCodeForTokens", nil)
		return nil, nil, nil, fmt.Errorf("authentication service is nil")
	}

	if req == nil {
		ec.Log.Error("ExchangeRequest is nil in exchangeCodeForTokens", nil)
		return nil, nil, nil, fmt.Errorf("exchange request is nil")
	}

	if req.Code == "" {
		ec.Log.Error("Authorization code is empty in exchangeCodeForTokens", nil)
		return nil, nil, nil, fmt.Errorf("authorization code is empty")
	}

	if req.State == "" {
		ec.Log.Error("State parameter is empty in exchangeCodeForTokens", nil)
		return nil, nil, nil, fmt.Errorf("state parameter is empty")
	}

	// Log parameters for debugging
	ec.Log.Debug("Calling authService.ExchangeCodeForTokens", map[string]any{
		"code_length":  len(req.Code),
		"state_length": len(req.State),
		"user_agent":   userAgent,
		"ip_address":   ipAddress,
		"service_nil":  authService == nil,
	})

	session, token, claims, err := authService.ExchangeCodeForTokens(req.Code, req.State, userAgent, ipAddress)
	if err != nil {
		ec.Log.Error("Failed to exchange code for tokens", map[string]any{
			"error": err.Error(),
			"code":  req.Code,
			"state": req.State,
		})
		return nil, nil, nil, err
	}
	return session, token, claims, nil
}

// setSessionCookie sets the secure session cookie in the HTTP response
func (ec *ExchangeController) setSessionCookie(c echo.Context, session *model.AuthSession) {
	if session == nil {
		ec.Log.Error("Session is nil in setSessionCookie", nil)
		return
	}

	if session.SessionID == "" {
		ec.Log.Error("SessionID is empty in setSessionCookie", nil)
		return
	}

	cookie := &http.Cookie{
		Name:     "app_session",
		Value:    session.SessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
		Expires:  session.RefreshExpiry,
	}
	c.SetCookie(cookie)
}

// buildResponse constructs the exchange response with token and user information
func (ec *ExchangeController) buildResponse(token *domain.Token, claims *domain.Claims) domain.ExchangeResponse {
	// Validate parameters
	if token == nil {
		ec.Log.Error("Token is nil in buildResponse", nil)
		return domain.ExchangeResponse{
			AccessToken: "",
			TokenType:   "Bearer",
			ExpiresIn:   0,
		}
	}

	if claims == nil {
		ec.Log.Error("Claims is nil in buildResponse", nil)
		return domain.ExchangeResponse{
			AccessToken: token.AccessToken,
			TokenType:   "Bearer",
			ExpiresIn:   0,
		}
	}

	// Calculate token expiry
	expiresIn := int64(0)
	if !token.Expiry.IsZero() {
		expiresIn = int64(time.Until(token.Expiry).Seconds())
	}

	response := domain.ExchangeResponse{
		AccessToken: token.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
	}

	// Set user information from claims
	response.User.ID = claims.Subject
	response.User.Username = claims.Username
	response.User.Name = claims.Name
	response.User.Email = claims.Email

	return response
}

// logSuccessfulExchange logs the successful token exchange operation
func (ec *ExchangeController) logSuccessfulExchange(claims *domain.Claims, session *model.AuthSession) {
	if claims == nil || session == nil {
		ec.Log.Error("Cannot log successful exchange: claims or session is nil", map[string]any{
			"claims_nil":  claims == nil,
			"session_nil": session == nil,
		})
		return
	}

	ec.Log.Info("Token exchange successful", map[string]any{
		"user_id":    claims.Subject,
		"username":   claims.Username,
		"session_id": session.SessionID,
	})
}
