package llmadmin

import (
	"context"

	"github.com/poly-workshop/llm-studio/internal/domain"
)

type SessionStore interface {
	Get(ctx context.Context, sessionID string) (domain.Session, error)
}

type UserRepository interface {
	GetRole(ctx context.Context, userID string) (domain.Role, error)
}

type LLMGatewayAdmin interface {
	UpsertProviderConfig(ctx context.Context, cfg domain.LLMProviderConfig) error
	DeleteProviderConfig(ctx context.Context, provider domain.LLMProviderType) error
	ListProviderConfigs(ctx context.Context) ([]domain.LLMProviderConfigView, error)

	UpsertModel(ctx context.Context, cfg domain.LLMModelConfig) (id string, err error)
	DeleteModel(ctx context.Context, id string) error
	ListModels(ctx context.Context) ([]domain.LLMModelSpec, error)
}
