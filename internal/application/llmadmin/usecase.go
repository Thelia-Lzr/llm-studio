package llmadmin

import (
	"context"
	"errors"
	"strings"

	"github.com/poly-workshop/llm-studio/internal/domain"
)

var (
	ErrUnauthenticated       = errors.New("unauthenticated")
	ErrForbidden             = errors.New("forbidden")
	ErrInvalidProvider       = errors.New("invalid provider")
	ErrProviderNotConfigured = errors.New("provider not configured")
	ErrInvalidModelID        = errors.New("invalid model id")
	ErrInvalidUpstreamModel  = errors.New("invalid upstream_model")
)

type Usecase struct {
	sessions SessionStore
	users    UserRepository
	admin    LLMGatewayAdmin
}

func New(admin LLMGatewayAdmin, sessions SessionStore, users UserRepository) *Usecase {
	return &Usecase{admin: admin, sessions: sessions, users: users}
}

func (u *Usecase) authorize(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return ErrUnauthenticated
	}
	sess, err := u.sessions.Get(ctx, sessionID)
	if err != nil {
		return ErrUnauthenticated
	}
	role, err := u.users.GetRole(ctx, sess.UserID)
	if err != nil {
		return err
	}
	if role != domain.RoleSuperAdmin && role != domain.RoleAdmin {
		return ErrForbidden
	}
	return nil
}

func (u *Usecase) ListProviderConfigs(ctx context.Context, sessionID string) ([]domain.LLMProviderConfigView, error) {
	if err := u.authorize(ctx, sessionID); err != nil {
		return nil, err
	}
	return u.admin.ListProviderConfigs(ctx)
}

func (u *Usecase) UpsertProviderConfig(ctx context.Context, sessionID string, cfg domain.LLMProviderConfig) error {
	if err := u.authorize(ctx, sessionID); err != nil {
		return err
	}
	if cfg.Provider == domain.LLMProviderUnspecified {
		return ErrInvalidProvider
	}
	cfg.BaseURL = strings.TrimSpace(cfg.BaseURL)
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)
	return u.admin.UpsertProviderConfig(ctx, cfg)
}

func (u *Usecase) DeleteProviderConfig(ctx context.Context, sessionID string, provider domain.LLMProviderType) error {
	if err := u.authorize(ctx, sessionID); err != nil {
		return err
	}
	if provider == domain.LLMProviderUnspecified {
		return ErrInvalidProvider
	}
	return u.admin.DeleteProviderConfig(ctx, provider)
}

func (u *Usecase) ListModels(ctx context.Context, sessionID string) ([]domain.LLMModelSpec, error) {
	if err := u.authorize(ctx, sessionID); err != nil {
		return nil, err
	}
	return u.admin.ListModels(ctx)
}

func (u *Usecase) UpsertModel(ctx context.Context, sessionID string, cfg domain.LLMModelConfig) (string, error) {
	if err := u.authorize(ctx, sessionID); err != nil {
		return "", err
	}
	if cfg.Provider == domain.LLMProviderUnspecified {
		return "", ErrInvalidProvider
	}
	// A model can only be configured for a provider that already has a provider config.
	providerConfigs, err := u.admin.ListProviderConfigs(ctx)
	if err != nil {
		return "", err
	}
	configured := false
	for _, pc := range providerConfigs {
		if pc.Provider == cfg.Provider {
			configured = true
			break
		}
	}
	if !configured {
		return "", ErrProviderNotConfigured
	}

	cfg.UpstreamModel = strings.TrimSpace(cfg.UpstreamModel)
	if cfg.UpstreamModel == "" {
		return "", ErrInvalidUpstreamModel
	}
	return u.admin.UpsertModel(ctx, cfg)
}

func (u *Usecase) DeleteModel(ctx context.Context, sessionID string, id string) error {
	if err := u.authorize(ctx, sessionID); err != nil {
		return err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrInvalidModelID
	}
	return u.admin.DeleteModel(ctx, id)
}
