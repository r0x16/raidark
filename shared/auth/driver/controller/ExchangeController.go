package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/r0x16/Raidark/shared/auth/domain"
	"github.com/r0x16/Raidark/shared/auth/domain/model"
	"github.com/r0x16/Raidark/shared/auth/driver/repositories"
	"github.com/r0x16/Raidark/shared/auth/service"
	domdatastore "github.com/r0x16/Raidark/shared/datastore/domain"
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
)

// ExchangeController handles OAuth2 authorization code exchange operations
type ExchangeController struct {
	Datastore domdatastore.DatabaseProvider
	Auth      domain.AuthProvider
	Log       domlogger.LogProvider
}

// ExchangeAction creates an ExchangeController instance and delegates to the Exchange method
func ExchangeAction(c echo.Context, hub *domprovider.ProviderHub) error {
	controller := &ExchangeController{
		Datastore: domprovider.Get[domdatastore.DatabaseProvider](hub),
		Auth:      domprovider.Get[domain.AuthProvider](hub),
		Log:       domprovider.Get[domlogger.LogProvider](hub),
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

	// Extract client information
	userAgent, ipAddress := ec.extractClientInfo(c)

	// Initialize services
	authService := ec.initializeAuthService(ec.Datastore)

	// Exchange code for tokens and create session
	session, token, claims, err := ec.exchangeCodeForTokens(authService, req, userAgent, ipAddress)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Authentication failed",
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
	sessionRepo := repositories.NewGormSessionRepository(dbProvider.GetDataStore().Exec)
	return service.NewAuthExchangeService(sessionRepo, ec.Auth)
}

// exchangeCodeForTokens performs the OAuth2 code exchange and returns session, token, and claims
func (ec *ExchangeController) exchangeCodeForTokens(authService *service.AuthExchangeService, req *domain.ExchangeRequest, userAgent, ipAddress string) (*model.AuthSession, *domain.Token, *domain.Claims, error) {
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
	ec.Log.Info("Token exchange successful", map[string]any{
		"user_id":    claims.Subject,
		"username":   claims.Username,
		"session_id": session.SessionID,
	})
}
