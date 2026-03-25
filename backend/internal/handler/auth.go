package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/newa/zeiterfassung/internal/config"
	"github.com/newa/zeiterfassung/internal/middleware"
	"github.com/newa/zeiterfassung/internal/service"
)

const sessionCookieName = "zeit_session"

// AuthHandler handles authentication HTTP endpoints.
type AuthHandler struct {
	authService *service.AuthService
	cfg         *config.Config
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService *service.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		cfg:         cfg,
	}
}

// Login initiates the OIDC login flow by redirecting to Azure AD.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	state, err := generateState()
	if err != nil {
		log.Error().Err(err).Msg("failed to generate OIDC state")
		ErrorJSON(w, http.StatusInternalServerError, "internal error")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.cfg.Env == "production",
	})

	authorizeURL := h.authService.AuthorizeURL(state)
	http.Redirect(w, r, authorizeURL, http.StatusTemporaryRedirect)
}

// Callback handles the OIDC callback from Azure AD.
func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		errDesc := r.URL.Query().Get("error_description")
		log.Warn().Str("error", errParam).Str("description", errDesc).Msg("OIDC error callback")
		http.Redirect(w, r, h.cfg.FrontendURL+"/login?error=auth_failed", http.StatusTemporaryRedirect)
		return
	}

	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		log.Warn().Msg("missing oauth_state cookie")
		http.Redirect(w, r, h.cfg.FrontendURL+"/login?error=invalid_state", http.StatusTemporaryRedirect)
		return
	}

	stateParam := r.URL.Query().Get("state")
	if stateParam == "" || stateParam != stateCookie.Value {
		log.Warn().Msg("OIDC state mismatch")
		http.Redirect(w, r, h.cfg.FrontendURL+"/login?error=invalid_state", http.StatusTemporaryRedirect)
		return
	}

	// Clear the state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		log.Warn().Msg("missing authorization code")
		http.Redirect(w, r, h.cfg.FrontendURL+"/login?error=no_code", http.StatusTemporaryRedirect)
		return
	}

	_, rawToken, err := h.authService.HandleCallback(r.Context(), code)
	if err != nil {
		log.Error().Err(err).Msg("OIDC callback failed")
		http.Redirect(w, r, h.cfg.FrontendURL+"/login?error=auth_failed", http.StatusTemporaryRedirect)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    rawToken,
		Path:     "/",
		MaxAge:   int(h.cfg.SessionMaxAge / time.Second),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.cfg.Env == "production",
	})

	http.Redirect(w, r, h.cfg.FrontendURL+"/dashboard", http.StatusTemporaryRedirect)
}

// Me returns the currently authenticated user's information.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		ErrorJSON(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	JSON(w, http.StatusOK, user)
}

// Logout invalidates the current session and clears the cookie.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil && cookie.Value != "" {
		if logoutErr := h.authService.Logout(r.Context(), cookie.Value); logoutErr != nil {
			log.Error().Err(logoutErr).Msg("failed to invalidate session")
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.cfg.Env == "production",
	})

	JSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

// generateState creates a random state parameter for CSRF protection in OIDC.
func generateState() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
