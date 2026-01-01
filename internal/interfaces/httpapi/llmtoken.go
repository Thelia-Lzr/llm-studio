package httpapi

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/poly-workshop/llm-studio/internal/application/llmtoken"
	infraConfig "github.com/poly-workshop/llm-studio/internal/infrastructure/config"
)

type llmTokenHandler struct {
	cfg infraConfig.Config
	uc  *llmtoken.Usecase
}

func newLLMTokenHandler(cfg infraConfig.Config, uc *llmtoken.Usecase) *llmTokenHandler {
	return &llmTokenHandler{cfg: cfg, uc: uc}
}

func (h *llmTokenHandler) issueToken(w http.ResponseWriter, r *http.Request) {
	sid, ok := h.sessionIDFromRequest(r)
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}

	ctx, cancel := contextWithTimeout(r, 5*time.Second)
	defer cancel()

	token, exp, err := h.uc.IssueToken(ctx, sid)
	if err != nil {
		switch err {
		case llmtoken.ErrUnauthenticated:
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
		default:
			http.Error(w, "failed", http.StatusInternalServerError)
		}
		return
	}

	// Store the data-plane JWT in an HttpOnly cookie so the browser can use it
	// without exposing the token to JS. The nginx proxy will forward it.
	maxAge := int(h.cfg.LLMGateway.TokenTTLSeconds)
	setLLMGatewayTokenCookie(w, h.cfg, token, maxAge)

	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"expires_at_unix": exp,
	})
}

func setLLMGatewayTokenCookie(w http.ResponseWriter, cfg infraConfig.Config, token string, maxAge int) {
	name := strings.TrimSpace(cfg.LLMGateway.AuthCookieName)
	if name == "" {
		name = "llmgw_access_token"
	}

	// Ensure cookie value is safe even if upstream starts returning tokens
	// with unexpected characters.
	value := token
	if !isCookieValueSafe(value) {
		value = base64.RawURLEncoding.EncodeToString([]byte(token))
	}

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

func isCookieValueSafe(v string) bool {
	// RFC6265 cookie-octet excludes CTLs, whitespace, DQUOTE, comma, semicolon, backslash.
	for i := 0; i < len(v); i++ {
		c := v[i]
		if c <= 0x20 || c >= 0x7f {
			return false
		}
		switch c {
		case '"', ',', ';', '\\':
			return false
		}
	}
	return true
}

func (h *llmTokenHandler) sessionIDFromRequest(r *http.Request) (string, bool) {
	c, err := r.Cookie(h.cfg.Auth.SessionCookieName)
	if err != nil || c.Value == "" {
		return "", false
	}
	return c.Value, true
}
