package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/poly-workshop/llm-studio/internal/application/rbac"
	"github.com/poly-workshop/llm-studio/internal/domain"
	infraConfig "github.com/poly-workshop/llm-studio/internal/infrastructure/config"
)

type rbacHandler struct {
	cfg  infraConfig.Config
	rbac *rbac.Usecase
}

func newRBACHandler(cfg infraConfig.Config, rbacUC *rbac.Usecase) *rbacHandler {
	return &rbacHandler{cfg: cfg, rbac: rbacUC}
}

func (h *rbacHandler) me(w http.ResponseWriter, r *http.Request) {
	sid, ok := h.sessionIDFromRequest(r)
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}
	me, err := h.rbac.Me(r.Context(), sid)
	if err != nil {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}

	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"user_id":   me.UserID,
		"role":      me.Role,
		"email":     me.Email,
		"github_id": me.GithubID,
		"nickname":  me.Nickname,
	})
}

type updateMeRequest struct {
	Nickname *string `json:"nickname"`
}

func (h *rbacHandler) updateMe(w http.ResponseWriter, r *http.Request) {
	sid, ok := h.sessionIDFromRequest(r)
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}

	var req updateMeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Nickname == nil {
		http.Error(w, "missing nickname", http.StatusBadRequest)
		return
	}

	me, err := h.rbac.UpdateMyNickname(r.Context(), sid, *req.Nickname)
	if err != nil {
		switch err {
		case rbac.ErrUnauthenticated:
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
		case rbac.ErrInvalidNickname:
			http.Error(w, "invalid nickname", http.StatusBadRequest)
		default:
			http.Error(w, "failed", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"user_id":   me.UserID,
		"role":      me.Role,
		"email":     me.Email,
		"github_id": me.GithubID,
		"nickname":  me.Nickname,
	})
}

type setRoleRequest struct {
	Role string `json:"role"`
}

func (h *rbacHandler) setUserRole(w http.ResponseWriter, r *http.Request) {
	sid, ok := h.sessionIDFromRequest(r)
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}

	targetUserID := strings.TrimSpace(r.PathValue("userID"))
	if targetUserID == "" {
		http.Error(w, "missing userID", http.StatusBadRequest)
		return
	}

	var req setRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	role, ok := domain.ParseRole(req.Role)
	if !ok {
		http.Error(w, "invalid role", http.StatusBadRequest)
		return
	}

	if err := h.rbac.SetUserRole(r.Context(), sid, targetUserID, role); err != nil {
		switch err {
		case rbac.ErrUnauthenticated:
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
		case rbac.ErrForbidden:
			http.Error(w, "forbidden", http.StatusForbidden)
		case rbac.ErrInvalidRole:
			http.Error(w, "invalid role", http.StatusBadRequest)
		default:
			http.Error(w, "failed", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *rbacHandler) listUsers(w http.ResponseWriter, r *http.Request) {
	sid, ok := h.sessionIDFromRequest(r)
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}

	limit := 50
	offset := 0
	if v := strings.TrimSpace(r.URL.Query().Get("limit")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	if v := strings.TrimSpace(r.URL.Query().Get("offset")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			offset = n
		}
	}

	users, err := h.rbac.ListUsers(r.Context(), sid, limit, offset)
	if err != nil {
		switch err {
		case rbac.ErrUnauthenticated:
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
		case rbac.ErrForbidden:
			http.Error(w, "forbidden", http.StatusForbidden)
		default:
			http.Error(w, "failed", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"users": users,
	})
}

func (h *rbacHandler) sessionIDFromRequest(r *http.Request) (string, bool) {
	c, err := r.Cookie(h.cfg.Auth.SessionCookieName)
	if err != nil || c.Value == "" {
		return "", false
	}
	return c.Value, true
}
