package rbac

import (
	"context"

	"github.com/poly-workshop/llm-studio/internal/domain"
)

type SessionStore interface {
	Get(ctx context.Context, sessionID string) (domain.Session, error)
}

type UserRepository interface {
	GetRole(ctx context.Context, userID string) (domain.Role, error)
	SetRole(ctx context.Context, userID string, role domain.Role) error
	ListUsers(ctx context.Context, limit int, offset int) ([]domain.AdminUser, error)
	GetMe(ctx context.Context, userID string) (domain.Me, error)
	SetNickname(ctx context.Context, userID string, nickname string) error
}
