package httpapi

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/poly-workshop/llm-studio/internal/application/auth"
	infraConfig "github.com/poly-workshop/llm-studio/internal/infrastructure/config"
)

type authHandler struct {
	cfg  infraConfig.Config
	auth *auth.Usecase
	now  func() time.Time
}

func newAuthHandler(cfg infraConfig.Config, authUC *auth.Usecase) *authHandler {
	return &authHandler{
		cfg:  cfg,
		auth: authUC,
		now:  time.Now,
	}
}

func (h *authHandler) githubLogin(w http.ResponseWriter, r *http.Request) {
	returnTo := strings.TrimSpace(r.URL.Query().Get("return_to"))
	if returnTo != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     h.cfg.Auth.ReturnToCookieName,
			Value:    url.QueryEscape(returnTo),
			Path:     h.cfg.Auth.CookiePath,
			Domain:   h.cfg.Auth.CookieDomain,
			Secure:   h.cfg.Auth.CookieSecure,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   h.cfg.Auth.ReturnToMaxAgeMins * 60,
		})
	}

	callbackURL := h.cfg.Public.BaseURL + "/api/auth/github/callback"

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	loginURL, _, err := h.auth.StartOAuth(ctx, h.cfg.Identra.OAuthProvider, callbackURL)
	if err != nil {
		http.Error(w, "failed to start oauth login", http.StatusBadGateway)
		return
	}

	http.Redirect(w, r, loginURL, http.StatusFound)
}

func (h *authHandler) githubCallback(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSpace(r.URL.Query().Get("code"))
	state := strings.TrimSpace(r.URL.Query().Get("state"))
	if code == "" || state == "" {
		http.Error(w, "missing code/state", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	ttl := time.Duration(h.cfg.Auth.CookieMaxAgeDays) * 24 * time.Hour
	sessionID, _, err := h.auth.CompleteOAuthToSession(ctx, code, state, ttl)
	if err != nil {
		http.Error(w, "oauth login failed", http.StatusUnauthorized)
		return
	}

	maxAge := h.cfg.Auth.CookieMaxAgeDays * 24 * 60 * 60
	setCookie(w, h.cfg, h.cfg.Auth.SessionCookieName, sessionID, maxAge)

	redirectPath := "/dashboard"
	if c, err := r.Cookie(h.cfg.Auth.ReturnToCookieName); err == nil && c.Value != "" {
		if v, err := url.QueryUnescape(c.Value); err == nil {
			if p := sanitizeReturnTo(v); p != "" {
				redirectPath = p
			}
		}
		http.SetCookie(w, &http.Cookie{
			Name:     h.cfg.Auth.ReturnToCookieName,
			Value:    "",
			Path:     h.cfg.Auth.CookiePath,
			Domain:   h.cfg.Auth.CookieDomain,
			Secure:   h.cfg.Auth.CookieSecure,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   -1,
		})
	}

	http.Redirect(w, r, h.cfg.Frontend.BaseURL+redirectPath, http.StatusFound)
}

func (h *authHandler) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(h.cfg.Auth.SessionCookieName); err == nil && c.Value != "" {
		_ = h.auth.Logout(r.Context(), c.Value)
	}
	clearCookie(w, h.cfg, h.cfg.Auth.SessionCookieName)
	w.WriteHeader(http.StatusNoContent)
}

func setCookie(w http.ResponseWriter, cfg infraConfig.Config, name, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     cfg.Auth.CookiePath,
		Domain:   cfg.Auth.CookieDomain,
		Secure:   cfg.Auth.CookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	})
}

func clearCookie(w http.ResponseWriter, cfg infraConfig.Config, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     cfg.Auth.CookiePath,
		Domain:   cfg.Auth.CookieDomain,
		Secure:   cfg.Auth.CookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func sanitizeReturnTo(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if strings.Contains(v, "://") || strings.HasPrefix(v, "//") {
		return ""
	}
	if !strings.HasPrefix(v, "/") {
		return ""
	}
	return v
}
