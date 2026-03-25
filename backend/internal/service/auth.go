package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/newa/zeiterfassung/internal/config"
	"github.com/newa/zeiterfassung/internal/model"
	"github.com/newa/zeiterfassung/internal/repository"
)

// AuthService handles authentication logic including OIDC and session management.
type AuthService struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
	cfg         *config.Config
}

// NewAuthService creates a new AuthService.
func NewAuthService(cfg *config.Config, userRepo *repository.UserRepository, sessionRepo *repository.SessionRepository) *AuthService {
	return &AuthService{
		cfg:         cfg,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

// OIDCTokenResponse represents the token endpoint response from Azure AD.
type OIDCTokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Error        string `json:"error,omitempty"`
	ErrorDesc    string `json:"error_description,omitempty"`
}

// MSGraphUserResponse represents the /me endpoint response from MS Graph.
type MSGraphUserResponse struct {
	ID                string `json:"id"`
	DisplayName       string `json:"displayName"`
	Mail              string `json:"mail"`
	UserPrincipalName string `json:"userPrincipalName"`
}

// AuthorizeURL returns the Azure AD authorization URL for initiating OIDC login.
func (s *AuthService) AuthorizeURL(state string) string {
	params := url.Values{
		"client_id":     {s.cfg.AzureAD.ClientID},
		"response_type": {"code"},
		"redirect_uri":  {s.cfg.AzureAD.RedirectURL},
		"scope":         {"openid profile email User.Read"},
		"state":         {state},
		"response_mode": {"query"},
	}
	return fmt.Sprintf(
		"https://login.microsoftonline.com/%s/oauth2/v2.0/authorize?%s",
		s.cfg.AzureAD.TenantID,
		params.Encode(),
	)
}

// ExchangeCode exchanges an authorization code for tokens via the Azure AD token endpoint.
func (s *AuthService) ExchangeCode(ctx context.Context, code string) (*OIDCTokenResponse, error) {
	tokenURL := fmt.Sprintf(
		"https://login.microsoftonline.com/%s/oauth2/v2.0/token",
		s.cfg.AzureAD.TenantID,
	)

	data := url.Values{
		"client_id":     {s.cfg.AzureAD.ClientID},
		"client_secret": {s.cfg.AzureAD.ClientSecret},
		"code":          {code},
		"redirect_uri":  {s.cfg.AzureAD.RedirectURL},
		"grant_type":    {"authorization_code"},
		"scope":         {"openid profile email User.Read"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token exchange request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("reading token response: %w", err)
	}

	var tokenResp OIDCTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("parsing token response: %w", err)
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("token error: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	return &tokenResp, nil
}

// FetchUserInfo fetches user information from the MS Graph /me endpoint.
func (s *AuthService) FetchUserInfo(ctx context.Context, accessToken string) (*MSGraphUserResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://graph.microsoft.com/v1.0/me", nil)
	if err != nil {
		return nil, fmt.Errorf("creating graph request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("graph request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("graph returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("reading graph response: %w", err)
	}

	var user MSGraphUserResponse
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("parsing graph response: %w", err)
	}

	return &user, nil
}

// HandleCallback processes the OIDC callback: exchanges code, fetches user info,
// upserts user, and creates a session. Returns the user, raw session token, and error.
func (s *AuthService) HandleCallback(ctx context.Context, code string) (*model.User, string, error) {
	tokenResp, err := s.ExchangeCode(ctx, code)
	if err != nil {
		return nil, "", fmt.Errorf("exchanging code: %w", err)
	}

	graphUser, err := s.FetchUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, "", fmt.Errorf("fetching user info: %w", err)
	}

	email := graphUser.Mail
	if email == "" {
		email = graphUser.UserPrincipalName
	}
	if email == "" {
		return nil, "", fmt.Errorf("no email found in user info")
	}

	user, err := s.userRepo.UpsertByAzureOID(ctx, email, graphUser.DisplayName, graphUser.ID)
	if err != nil {
		return nil, "", fmt.Errorf("upserting user: %w", err)
	}

	if !user.IsActive {
		return nil, "", fmt.Errorf("user account is deactivated")
	}

	rawToken, err := generateSessionToken()
	if err != nil {
		return nil, "", fmt.Errorf("generating session token: %w", err)
	}

	tokenHash := hashToken(rawToken)
	expiresAt := time.Now().Add(s.cfg.SessionMaxAge)

	_, err = s.sessionRepo.Create(ctx, user.ID, tokenHash, expiresAt)
	if err != nil {
		return nil, "", fmt.Errorf("creating session: %w", err)
	}

	log.Info().Str("user_id", user.ID.String()).Str("email", user.Email).Msg("user authenticated via OIDC")

	return user, rawToken, nil
}

// ValidateSession checks if a session token is valid and returns the associated user.
func (s *AuthService) ValidateSession(ctx context.Context, rawToken string) (*model.User, error) {
	tokenHash := hashToken(rawToken)

	session, err := s.sessionRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("finding session: %w", err)
	}
	if session == nil {
		return nil, nil
	}

	user, err := s.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}
	if user == nil || !user.IsActive {
		return nil, nil
	}

	return user, nil
}

// Logout invalidates all sessions for a user.
func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	tokenHash := hashToken(rawToken)

	session, err := s.sessionRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("finding session: %w", err)
	}
	if session == nil {
		return nil
	}

	return s.sessionRepo.DeleteByUserID(ctx, session.UserID)
}

// generateSessionToken creates a cryptographically random session token.
func generateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// hashToken produces a SHA-256 hash of a raw token for storage.
func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
