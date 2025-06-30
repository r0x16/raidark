package auth

import (
	"fmt"
	"net/url"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	domauth "github.com/r0x16/Raidark/shared/domain/auth"
	"github.com/r0x16/Raidark/shared/domain/model/auth"
	"golang.org/x/oauth2"
)

// CasdoorAuthProvider implements the AuthProvider interface using Casdoor
type CasdoorAuthProvider struct {
	config *CasdoorConfig
	client *casdoorsdk.Client
}

// Verify interface implementation
var _ domauth.AuthProvider = &CasdoorAuthProvider{}

// NewCasdoorAuthProvider creates a new CasdoorAuthProvider instance
func NewCasdoorAuthProvider(config *CasdoorConfig) *CasdoorAuthProvider {
	return &CasdoorAuthProvider{
		config: config,
	}
}

// NewCasdoorAuthProviderFromEnv creates a new CasdoorAuthProvider from environment variables
func NewCasdoorAuthProviderFromEnv() *CasdoorAuthProvider {
	config := NewCasdoorConfigFromEnv()
	return NewCasdoorAuthProvider(config)
}

// Initialize the auth provider with configuration
func (c *CasdoorAuthProvider) Initialize() error {
	if err := c.config.Validate(); err != nil {
		return newCasdoorErrorWithCause("failed to validate configuration", err)
	}

	// Initialize the Casdoor client
	c.client = casdoorsdk.NewClient(
		c.config.Endpoint,
		c.config.ClientId,
		c.config.ClientSecret,
		c.config.Certificate,
		c.config.OrganizationName,
		c.config.ApplicationName,
	)

	return nil
}

// GetAuthURL gets OAuth authorization URL for user login
func (c *CasdoorAuthProvider) GetAuthURL(state string) string {
	if c.client == nil {
		return ""
	}

	// Build OAuth authorization URL manually according to Casdoor documentation
	authURL := fmt.Sprintf("%s/login/oauth/authorize", c.config.Endpoint)
	params := url.Values{}
	params.Add("client_id", c.config.ClientId)
	params.Add("redirect_uri", c.config.RedirectURI)
	params.Add("response_type", "code")
	params.Add("scope", "openid profile email")
	params.Add("state", state)

	return fmt.Sprintf("%s?%s", authURL, params.Encode())
}

// GetToken exchanges authorization code for access token
func (c *CasdoorAuthProvider) GetToken(code, state string) (*auth.Token, error) {
	if c.client == nil {
		return nil, newCasdoorError("client not initialized")
	}

	oauthToken, err := c.client.GetOAuthToken(code, state)
	if err != nil {
		return nil, newCasdoorErrorWithCause("failed to get OAuth token", err)
	}

	return c.convertOAuth2TokenToDomainToken(oauthToken), nil
}

// RefreshToken refreshes OAuth token using refresh token
func (c *CasdoorAuthProvider) RefreshToken(refreshToken string) (*auth.Token, error) {
	if c.client == nil {
		return nil, newCasdoorError("client not initialized")
	}

	oauthToken, err := c.client.RefreshOAuthToken(refreshToken)
	if err != nil {
		return nil, newCasdoorErrorWithCause("failed to refresh OAuth token", err)
	}

	return c.convertOAuth2TokenToDomainToken(oauthToken), nil
}

// ParseToken parses and validates JWT token
func (c *CasdoorAuthProvider) ParseToken(token string) (*auth.Claims, error) {
	if c.client == nil {
		return nil, newCasdoorError("client not initialized")
	}

	casdoorClaims, err := c.client.ParseJwtToken(token)
	if err != nil {
		return nil, newCasdoorErrorWithCause("failed to parse JWT token", err)
	}

	return c.convertCasdoorClaimsToDomainClaims(casdoorClaims), nil
}

// GetUser gets user information by username
func (c *CasdoorAuthProvider) GetUser(username string) (*auth.User, error) {
	if c.client == nil {
		return nil, newCasdoorError("client not initialized")
	}

	casdoorUser, err := c.client.GetUser(username)
	if err != nil {
		return nil, newCasdoorErrorWithCause(fmt.Sprintf("failed to get user: %s", username), err)
	}

	return c.convertCasdoorUserToDomainUser(casdoorUser), nil
}

// GetUsers gets all users
func (c *CasdoorAuthProvider) GetUsers() ([]*auth.User, error) {
	if c.client == nil {
		return nil, newCasdoorError("client not initialized")
	}

	casdoorUsers, err := c.client.GetUsers()
	if err != nil {
		return nil, newCasdoorErrorWithCause("failed to get users", err)
	}

	domainUsers := make([]*auth.User, len(casdoorUsers))
	for i, casdoorUser := range casdoorUsers {
		domainUsers[i] = c.convertCasdoorUserToDomainUser(casdoorUser)
	}

	return domainUsers, nil
}

// AddUser creates a new user
func (c *CasdoorAuthProvider) AddUser(user *auth.User) (bool, error) {
	if c.client == nil {
		return false, newCasdoorError("client not initialized")
	}

	casdoorUser := c.convertDomainUserToCasdoorUser(user)
	success, err := c.client.AddUser(casdoorUser)
	if err != nil {
		return false, newCasdoorErrorWithCause("failed to add user", err)
	}

	return success, nil
}

// UpdateUser updates existing user
func (c *CasdoorAuthProvider) UpdateUser(user *auth.User) (bool, error) {
	if c.client == nil {
		return false, newCasdoorError("client not initialized")
	}

	casdoorUser := c.convertDomainUserToCasdoorUser(user)
	success, err := c.client.UpdateUser(casdoorUser)
	if err != nil {
		return false, newCasdoorErrorWithCause("failed to update user", err)
	}

	return success, nil
}

// DeleteUser deletes user
func (c *CasdoorAuthProvider) DeleteUser(user *auth.User) (bool, error) {
	if c.client == nil {
		return false, newCasdoorError("client not initialized")
	}

	casdoorUser := c.convertDomainUserToCasdoorUser(user)
	success, err := c.client.DeleteUser(casdoorUser)
	if err != nil {
		return false, newCasdoorErrorWithCause("failed to delete user", err)
	}

	return success, nil
}

// HealthCheck verifies if provider is healthy
func (c *CasdoorAuthProvider) HealthCheck() error {
	if c.client == nil {
		return newCasdoorError("client not initialized")
	}

	// Try to get users as a simple health check
	_, err := c.client.GetUsers()
	if err != nil {
		return newCasdoorErrorWithCause("health check failed", err)
	}

	return nil
}

// Conversion functions

func (c *CasdoorAuthProvider) convertOAuth2TokenToDomainToken(token *oauth2.Token) *auth.Token {
	return &auth.Token{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}
}

func (c *CasdoorAuthProvider) convertCasdoorClaimsToDomainClaims(claims *casdoorsdk.Claims) *auth.Claims {
	domainClaims := &auth.Claims{
		Username:     claims.User.Name,
		Name:         claims.User.DisplayName,
		Email:        claims.User.Email,
		Organization: claims.User.Owner,
		Type:         claims.User.Type,
		Issuer:       claims.Issuer,
		Subject:      claims.Subject,
	}

	// Handle Audience - convert slice to string
	if len(claims.Audience) > 0 {
		domainClaims.Audience = claims.Audience[0]
	}

	// Handle NumericDate fields
	if claims.ExpiresAt != nil {
		domainClaims.ExpiresAt = claims.ExpiresAt.Unix()
	}
	if claims.IssuedAt != nil {
		domainClaims.IssuedAt = claims.IssuedAt.Unix()
	}
	if claims.NotBefore != nil {
		domainClaims.NotBefore = claims.NotBefore.Unix()
	}

	return domainClaims
}

func (c *CasdoorAuthProvider) convertCasdoorUserToDomainUser(user *casdoorsdk.User) *auth.User {
	domainUser := &auth.User{
		Owner:               user.Owner,
		Name:                user.Name,
		CreatedTime:         user.CreatedTime,
		UpdatedTime:         user.UpdatedTime,
		ID:                  user.Id,
		Type:                user.Type,
		Password:            user.Password,
		PasswordSalt:        user.PasswordSalt,
		DisplayName:         user.DisplayName,
		FirstName:           user.FirstName,
		LastName:            user.LastName,
		Avatar:              user.Avatar,
		PermanentAvatar:     user.PermanentAvatar,
		Email:               user.Email,
		EmailVerified:       user.EmailVerified,
		Phone:               user.Phone,
		PhoneVerified:       false, // Field not available in casdoorsdk.User
		CountryCode:         user.CountryCode,
		Region:              user.Region,
		Location:            user.Location,
		Address:             user.Address,
		Affiliation:         user.Affiliation,
		Title:               user.Title,
		IdCardType:          user.IdCardType,
		IdCard:              user.IdCard,
		Homepage:            user.Homepage,
		Bio:                 user.Bio,
		Language:            user.Language,
		Gender:              user.Gender,
		Birthday:            user.Birthday,
		Education:           user.Education,
		Score:               user.Score,
		Karma:               user.Karma,
		Ranking:             user.Ranking,
		IsOnline:            user.IsOnline,
		IsAdmin:             user.IsAdmin,
		IsGlobalAdmin:       false, // Field not available in casdoorsdk.User
		IsForbidden:         user.IsForbidden,
		IsDeleted:           user.IsDeleted,
		SignupApplication:   user.SignupApplication,
		Hash:                user.Hash,
		PreHash:             user.PreHash,
		CreatedIP:           user.CreatedIp,
		LastSigninTime:      user.LastSigninTime,
		LastSigninIP:        user.LastSigninIp,
		GitHub:              user.GitHub,
		Google:              user.Google,
		QQ:                  user.QQ,
		WeChat:              user.WeChat,
		Facebook:            user.Facebook,
		DingTalk:            user.DingTalk,
		Weibo:               user.Weibo,
		Gitee:               user.Gitee,
		LinkedIn:            user.LinkedIn,
		Wecom:               user.Wecom,
		Lark:                user.Lark,
		Gitlab:              user.Gitlab,
		ADFS:                user.Adfs,
		Baidu:               user.Baidu,
		Alipay:              user.Alipay,
		Casdoor:             user.Casdoor,
		Infoflow:            user.Infoflow,
		Apple:               user.Apple,
		AzureAD:             user.AzureAD,
		Slack:               user.Slack,
		Steam:               user.Steam,
		Bilibili:            user.Bilibili,
		Okta:                user.Okta,
		Douyin:              user.Douyin,
		Line:                user.Line,
		Amazon:              user.Amazon,
		Instagram:           user.Instagram,
		TikTok:              user.TikTok,
		Dropbox:             user.Dropbox,
		Yahoo:               user.Yahoo,
		Yandex:              user.Yandex,
		StackOverflow:       "", // Field not available in casdoorsdk.User
		Tumblr:              user.Tumblr,
		Mailru:              user.Mailru,
		Battle:              "", // Field not available in casdoorsdk.User
		Uber:                user.Uber,
		NextCloud:           user.Nextcloud,
		Naver:               user.Naver,
		Kakao:               user.Kakao,
		VK:                  user.VK,
		Patreon:             user.Patreon,
		Custom:              user.Custom,
		Ldap:                user.Ldap,
		Properties:          user.Properties,
		Groups:              user.Groups,
		LastSigninWrongTime: user.LastSigninWrongTime,
		SigninWrongTimes:    user.SigninWrongTimes,
		Tag:                 user.Tag,
	}

	// Convert roles
	if user.Roles != nil {
		domainUser.Roles = make([]*auth.Role, len(user.Roles))
		for i, role := range user.Roles {
			domainUser.Roles[i] = &auth.Role{
				Owner:       role.Owner,
				Name:        role.Name,
				CreatedTime: role.CreatedTime,
				DisplayName: role.DisplayName,
				Description: role.Description,
				Users:       role.Users,
				Groups:      role.Groups,
				Domains:     role.Domains,
			}
		}
	}

	// Convert permissions
	if user.Permissions != nil {
		domainUser.Permissions = make([]*auth.Permission, len(user.Permissions))
		for i, perm := range user.Permissions {
			domainUser.Permissions[i] = &auth.Permission{
				Owner:        perm.Owner,
				Name:         perm.Name,
				CreatedTime:  perm.CreatedTime,
				DisplayName:  perm.DisplayName,
				Description:  perm.Description,
				Users:        perm.Users,
				Groups:       perm.Groups,
				Roles:        perm.Roles,
				Domains:      perm.Domains,
				Model:        perm.Model,
				Adapter:      perm.Adapter,
				ResourceType: perm.ResourceType,
				Resources:    perm.Resources,
				Actions:      perm.Actions,
				Effect:       perm.Effect,
				IsEnabled:    perm.IsEnabled,
				Submitter:    perm.Submitter,
				Approver:     perm.Approver,
				ApproveTime:  perm.ApproveTime,
				State:        perm.State,
			}
		}
	}

	// Convert managed accounts
	if user.ManagedAccounts != nil {
		domainUser.ManagedAccounts = make([]auth.ManagedAccount, len(user.ManagedAccounts))
		for i, account := range user.ManagedAccounts {
			domainUser.ManagedAccounts[i] = auth.ManagedAccount{
				Application: account.Application,
				Username:    account.Username,
				Password:    account.Password,
				SigninURL:   account.SigninUrl,
			}
		}
	}

	return domainUser
}

func (c *CasdoorAuthProvider) convertDomainUserToCasdoorUser(user *auth.User) *casdoorsdk.User {
	casdoorUser := &casdoorsdk.User{
		Owner:               user.Owner,
		Name:                user.Name,
		CreatedTime:         user.CreatedTime,
		UpdatedTime:         user.UpdatedTime,
		Id:                  user.ID,
		Type:                user.Type,
		Password:            user.Password,
		PasswordSalt:        user.PasswordSalt,
		DisplayName:         user.DisplayName,
		FirstName:           user.FirstName,
		LastName:            user.LastName,
		Avatar:              user.Avatar,
		PermanentAvatar:     user.PermanentAvatar,
		Email:               user.Email,
		EmailVerified:       user.EmailVerified,
		Phone:               user.Phone,
		CountryCode:         user.CountryCode,
		Region:              user.Region,
		Location:            user.Location,
		Address:             user.Address,
		Affiliation:         user.Affiliation,
		Title:               user.Title,
		IdCardType:          user.IdCardType,
		IdCard:              user.IdCard,
		Homepage:            user.Homepage,
		Bio:                 user.Bio,
		Language:            user.Language,
		Gender:              user.Gender,
		Birthday:            user.Birthday,
		Education:           user.Education,
		Score:               user.Score,
		Karma:               user.Karma,
		Ranking:             user.Ranking,
		IsOnline:            user.IsOnline,
		IsAdmin:             user.IsAdmin,
		IsForbidden:         user.IsForbidden,
		IsDeleted:           user.IsDeleted,
		SignupApplication:   user.SignupApplication,
		Hash:                user.Hash,
		PreHash:             user.PreHash,
		CreatedIp:           user.CreatedIP,
		LastSigninTime:      user.LastSigninTime,
		LastSigninIp:        user.LastSigninIP,
		GitHub:              user.GitHub,
		Google:              user.Google,
		QQ:                  user.QQ,
		WeChat:              user.WeChat,
		Facebook:            user.Facebook,
		DingTalk:            user.DingTalk,
		Weibo:               user.Weibo,
		Gitee:               user.Gitee,
		LinkedIn:            user.LinkedIn,
		Wecom:               user.Wecom,
		Lark:                user.Lark,
		Gitlab:              user.Gitlab,
		Adfs:                user.ADFS,
		Baidu:               user.Baidu,
		Alipay:              user.Alipay,
		Casdoor:             user.Casdoor,
		Infoflow:            user.Infoflow,
		Apple:               user.Apple,
		AzureAD:             user.AzureAD,
		Slack:               user.Slack,
		Steam:               user.Steam,
		Bilibili:            user.Bilibili,
		Okta:                user.Okta,
		Douyin:              user.Douyin,
		Line:                user.Line,
		Amazon:              user.Amazon,
		Instagram:           user.Instagram,
		TikTok:              user.TikTok,
		Dropbox:             user.Dropbox,
		Yahoo:               user.Yahoo,
		Yandex:              user.Yandex,
		Tumblr:              user.Tumblr,
		Mailru:              user.Mailru,
		Uber:                user.Uber,
		Nextcloud:           user.NextCloud,
		Naver:               user.Naver,
		Kakao:               user.Kakao,
		VK:                  user.VK,
		Patreon:             user.Patreon,
		Custom:              user.Custom,
		Ldap:                user.Ldap,
		Properties:          user.Properties,
		Groups:              user.Groups,
		LastSigninWrongTime: user.LastSigninWrongTime,
		SigninWrongTimes:    user.SigninWrongTimes,
		Tag:                 user.Tag,
	}

	// Convert roles
	if user.Roles != nil {
		casdoorUser.Roles = make([]*casdoorsdk.Role, len(user.Roles))
		for i, role := range user.Roles {
			casdoorUser.Roles[i] = &casdoorsdk.Role{
				Owner:       role.Owner,
				Name:        role.Name,
				CreatedTime: role.CreatedTime,
				DisplayName: role.DisplayName,
				Description: role.Description,
				Users:       role.Users,
				Groups:      role.Groups,
				Domains:     role.Domains,
			}
		}
	}

	// Convert permissions
	if user.Permissions != nil {
		casdoorUser.Permissions = make([]*casdoorsdk.Permission, len(user.Permissions))
		for i, perm := range user.Permissions {
			casdoorUser.Permissions[i] = &casdoorsdk.Permission{
				Owner:        perm.Owner,
				Name:         perm.Name,
				CreatedTime:  perm.CreatedTime,
				DisplayName:  perm.DisplayName,
				Description:  perm.Description,
				Users:        perm.Users,
				Groups:       perm.Groups,
				Roles:        perm.Roles,
				Domains:      perm.Domains,
				Model:        perm.Model,
				Adapter:      perm.Adapter,
				ResourceType: perm.ResourceType,
				Resources:    perm.Resources,
				Actions:      perm.Actions,
				Effect:       perm.Effect,
				IsEnabled:    perm.IsEnabled,
				Submitter:    perm.Submitter,
				Approver:     perm.Approver,
				ApproveTime:  perm.ApproveTime,
				State:        perm.State,
			}
		}
	}

	// Convert managed accounts
	if user.ManagedAccounts != nil {
		accounts := make([]casdoorsdk.ManagedAccount, len(user.ManagedAccounts))
		for i, account := range user.ManagedAccounts {
			accounts[i] = casdoorsdk.ManagedAccount{
				Application: account.Application,
				Username:    account.Username,
				Password:    account.Password,
				SigninUrl:   account.SigninURL,
			}
		}
		casdoorUser.ManagedAccounts = accounts
	}

	return casdoorUser
}
