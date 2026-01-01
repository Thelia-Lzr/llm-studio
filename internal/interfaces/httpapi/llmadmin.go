package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/poly-workshop/llm-studio/internal/application/llmadmin"
	"github.com/poly-workshop/llm-studio/internal/domain"
	infraConfig "github.com/poly-workshop/llm-studio/internal/infrastructure/config"
)

type llmAdminHandler struct {
	cfg infraConfig.Config
	uc  *llmadmin.Usecase
}

func newLLMAdminHandler(cfg infraConfig.Config, uc *llmadmin.Usecase) *llmAdminHandler {
	return &llmAdminHandler{cfg: cfg, uc: uc}
}

func (h *llmAdminHandler) listProviders(w http.ResponseWriter, r *http.Request) {
	sid, ok := h.sessionIDFromRequest(r)
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}

	configs, err := h.uc.ListProviderConfigs(r.Context(), sid)
	if err != nil {
		switch err {
		case llmadmin.ErrUnauthenticated:
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
		case llmadmin.ErrForbidden:
			http.Error(w, "forbidden", http.StatusForbidden)
		default:
			http.Error(w, "failed", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"configs": configs})
}

type upsertProviderRequest struct {
	BaseURL        string `json:"base_url"`
	APIKey         string `json:"api_key"`
	TimeoutSeconds *int64 `json:"timeout_seconds"`
}

func (h *llmAdminHandler) upsertProvider(w http.ResponseWriter, r *http.Request) {
	sid, ok := h.sessionIDFromRequest(r)
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}

	providerStr := strings.TrimSpace(r.PathValue("provider"))
	provider, ok := domain.ParseLLMProviderType(providerStr)
	if !ok {
		http.Error(w, "invalid provider", http.StatusBadRequest)
		return
	}

	var req upsertProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	cfg := domain.LLMProviderConfig{
		Provider: provider,
		BaseURL:  strings.TrimSpace(req.BaseURL),
		APIKey:   strings.TrimSpace(req.APIKey),
	}
	if req.TimeoutSeconds != nil {
		cfg.TimeoutSeconds = *req.TimeoutSeconds
	}

	if err := h.uc.UpsertProviderConfig(r.Context(), sid, cfg); err != nil {
		switch err {
		case llmadmin.ErrUnauthenticated:
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
		case llmadmin.ErrForbidden:
			http.Error(w, "forbidden", http.StatusForbidden)
		case llmadmin.ErrInvalidProvider:
			http.Error(w, "invalid provider", http.StatusBadRequest)
		default:
			http.Error(w, "failed", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *llmAdminHandler) deleteProvider(w http.ResponseWriter, r *http.Request) {
	sid, ok := h.sessionIDFromRequest(r)
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}

	providerStr := strings.TrimSpace(r.PathValue("provider"))
	provider, ok := domain.ParseLLMProviderType(providerStr)
	if !ok {
		http.Error(w, "invalid provider", http.StatusBadRequest)
		return
	}

	if err := h.uc.DeleteProviderConfig(r.Context(), sid, provider); err != nil {
		switch err {
		case llmadmin.ErrUnauthenticated:
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
		case llmadmin.ErrForbidden:
			http.Error(w, "forbidden", http.StatusForbidden)
		case llmadmin.ErrInvalidProvider:
			http.Error(w, "invalid provider", http.StatusBadRequest)
		default:
			http.Error(w, "failed", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type upsertModelRequest struct {
	Provider      string   `json:"provider"`
	UpstreamModel string   `json:"upstream_model"`
	Capabilities  []string `json:"capabilities"`
}

func (h *llmAdminHandler) listModels(w http.ResponseWriter, r *http.Request) {
	sid, ok := h.sessionIDFromRequest(r)
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}

	models, err := h.uc.ListModels(r.Context(), sid)
	if err != nil {
		switch err {
		case llmadmin.ErrUnauthenticated:
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
		case llmadmin.ErrForbidden:
			http.Error(w, "forbidden", http.StatusForbidden)
		default:
			http.Error(w, "failed", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"models": models})
}

func (h *llmAdminHandler) upsertModel(w http.ResponseWriter, r *http.Request) {
	sid, ok := h.sessionIDFromRequest(r)
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}

	var req upsertModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	provider, ok := domain.ParseLLMProviderType(req.Provider)
	if !ok {
		http.Error(w, "invalid provider", http.StatusBadRequest)
		return
	}

	caps := make([]domain.LLMModelCapability, 0, len(req.Capabilities))
	for _, s := range req.Capabilities {
		cap, ok := domain.ParseLLMModelCapability(s)
		if !ok {
			http.Error(w, "invalid capability", http.StatusBadRequest)
			return
		}
		caps = append(caps, cap)
	}

	id, err := h.uc.UpsertModel(r.Context(), sid, domain.LLMModelConfig{
		Provider:      provider,
		UpstreamModel: req.UpstreamModel,
		Capabilities:  caps,
	})
	if err != nil {
		switch err {
		case llmadmin.ErrUnauthenticated:
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
		case llmadmin.ErrForbidden:
			http.Error(w, "forbidden", http.StatusForbidden)
		case llmadmin.ErrInvalidProvider, llmadmin.ErrInvalidUpstreamModel, llmadmin.ErrProviderNotConfigured:
			http.Error(w, "invalid request", http.StatusBadRequest)
		default:
			http.Error(w, "failed", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"id": id})
}

func (h *llmAdminHandler) deleteModel(w http.ResponseWriter, r *http.Request) {
	sid, ok := h.sessionIDFromRequest(r)
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}

	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	if err := h.uc.DeleteModel(r.Context(), sid, id); err != nil {
		switch err {
		case llmadmin.ErrUnauthenticated:
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
		case llmadmin.ErrForbidden:
			http.Error(w, "forbidden", http.StatusForbidden)
		case llmadmin.ErrInvalidModelID:
			http.Error(w, "invalid id", http.StatusBadRequest)
		default:
			http.Error(w, "failed", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *llmAdminHandler) sessionIDFromRequest(r *http.Request) (string, bool) {
	c, err := r.Cookie(h.cfg.Auth.SessionCookieName)
	if err != nil || c.Value == "" {
		return "", false
	}
	return c.Value, true
}
