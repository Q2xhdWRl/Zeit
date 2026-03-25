package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/newa/zeiterfassung/internal/service"
)

func newTestAuthHandler() *AuthHandler {
	cfg := testConfig
	authService := service.NewAuthService(&cfg, nil, nil)
	return NewAuthHandler(authService, &cfg)
}

func TestLogin_RedirectsToAzureAD(t *testing.T) {
	h := newTestAuthHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected status %d, got %d", http.StatusTemporaryRedirect, rec.Code)
	}

	location := rec.Header().Get("Location")
	if location == "" {
		t.Error("expected Location header to be set")
	}

	// Should set the oauth_state cookie
	cookies := rec.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "oauth_state" {
			found = true
			if c.Value == "" {
				t.Error("oauth_state cookie should not be empty")
			}
		}
	}
	if !found {
		t.Error("expected oauth_state cookie to be set")
	}
}

func TestCallback_MissingCode_RedirectsWithError(t *testing.T) {
	h := newTestAuthHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/auth/callback?state=abc", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "abc"})
	rec := httptest.NewRecorder()

	h.Callback(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected status %d, got %d", http.StatusTemporaryRedirect, rec.Code)
	}

	location := rec.Header().Get("Location")
	expected := "http://localhost:3000/login?error=no_code"
	if location != expected {
		t.Errorf("expected redirect to %s, got %s", expected, location)
	}
}

func TestCallback_StateMismatch_RedirectsWithError(t *testing.T) {
	h := newTestAuthHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/auth/callback?state=wrong", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "correct"})
	rec := httptest.NewRecorder()

	h.Callback(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected status %d, got %d", http.StatusTemporaryRedirect, rec.Code)
	}

	location := rec.Header().Get("Location")
	expected := "http://localhost:3000/login?error=invalid_state"
	if location != expected {
		t.Errorf("expected redirect to %s, got %s", expected, location)
	}
}

func TestCallback_OIDCError_RedirectsWithError(t *testing.T) {
	h := newTestAuthHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/auth/callback?error=access_denied&error_description=test", nil)
	rec := httptest.NewRecorder()

	h.Callback(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected status %d, got %d", http.StatusTemporaryRedirect, rec.Code)
	}

	location := rec.Header().Get("Location")
	expected := "http://localhost:3000/login?error=auth_failed"
	if location != expected {
		t.Errorf("expected redirect to %s, got %s", expected, location)
	}
}

func TestMe_Unauthenticated_Returns401(t *testing.T) {
	h := newTestAuthHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rec := httptest.NewRecorder()

	h.Me(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestLogout_ClearsCookie(t *testing.T) {
	h := newTestAuthHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	rec := httptest.NewRecorder()

	h.Logout(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	cookies := rec.Result().Cookies()
	for _, c := range cookies {
		if c.Name == sessionCookieName && c.MaxAge != -1 {
			t.Error("expected session cookie to be cleared (MaxAge=-1)")
		}
	}
}
