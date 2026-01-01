package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/poly-workshop/llm-studio/internal/application/auth"
	"github.com/poly-workshop/llm-studio/internal/application/llmadmin"
	"github.com/poly-workshop/llm-studio/internal/application/llmtoken"
	"github.com/poly-workshop/llm-studio/internal/application/rbac"
	infraConfig "github.com/poly-workshop/llm-studio/internal/infrastructure/config"
)

func NewMux(cfg infraConfig.Config, authUC *auth.Usecase, rbacUC *rbac.Usecase, llmAdminUC *llmadmin.Usecase, llmTokenUC *llmtoken.Usecase) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	authH := newAuthHandler(cfg, authUC)
	mux.HandleFunc("GET /api/auth/github/login", authH.githubLogin)
	mux.HandleFunc("GET /api/auth/github/callback", authH.githubCallback)
	mux.HandleFunc("POST /api/auth/logout", authH.logout)

	rbacH := newRBACHandler(cfg, rbacUC)
	mux.HandleFunc("GET /api/me", rbacH.me)
	mux.HandleFunc("PATCH /api/me", rbacH.updateMe)
	mux.HandleFunc("GET /api/admin/users", rbacH.listUsers)
	mux.HandleFunc("POST /api/admin/users/{userID}/role", rbacH.setUserRole)

	llmH := newLLMAdminHandler(cfg, llmAdminUC)
	mux.HandleFunc("GET /api/admin/llm/providers", llmH.listProviders)
	mux.HandleFunc("PUT /api/admin/llm/providers/{provider}", llmH.upsertProvider)
	mux.HandleFunc("DELETE /api/admin/llm/providers/{provider}", llmH.deleteProvider)
	mux.HandleFunc("GET /api/admin/llm/models", llmH.listModels)
	mux.HandleFunc("POST /api/admin/llm/models", llmH.upsertModel)
	mux.HandleFunc("DELETE /api/admin/llm/models", llmH.deleteModel)

	llmTokenH := newLLMTokenHandler(cfg, llmTokenUC)
	mux.HandleFunc("POST /api/llm/token", llmTokenH.issueToken)

	return mux
}

func contextWithTimeout(r *http.Request, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), d)
}
