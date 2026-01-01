package auth

import (
	"context"
	"time"

	"github.com/poly-workshop/llm-studio/internal/domain"
)

// OAuthGateway is the port for external identity providers (implemented by infrastructure, e.g. identra gRPC).
type OAuthGateway interface {
	GetAuthorizationURL(ctx context.Context, provider string, redirectURL string) (url string, state string, err error)
	LoginByOAuth(ctx context.Context, code string, state string) (domain.TokenPair, error)
	GetCurrentUserLoginInfo(ctx context.Context, accessToken string) (domain.LoginInfo, error)
}

// UIDExtractor extracts identra uid from an access token.
type UIDExtractor interface {
	ExtractUserID(accessToken string) (string, error)
}

// SessionStore is the port for persisting sessions (implemented by infrastructure, e.g. Redis).
type SessionStore interface {
	Save(ctx context.Context, sessionID string, session domain.Session, ttl time.Duration) error
	Delete(ctx context.Context, sessionID string) error
}

// UserRepository persists local users (primary key = identra uid).
type UserRepository interface {
	EnsureExists(ctx context.Context, userID string) error
	SaveLoginInfo(ctx context.Context, info domain.LoginInfo) error
	GetRole(ctx context.Context, userID string) (domain.Role, error)
	SetRole(ctx context.Context, userID string, role domain.Role) error
}
