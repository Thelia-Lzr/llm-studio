package llmtoken

import (
	"context"

	"github.com/poly-workshop/llm-studio/internal/domain"
)

type SessionStore interface {
	Get(ctx context.Context, sessionID string) (domain.Session, error)
}

type TokenIssuer interface {
	IssueToken(ctx context.Context, subject string, ttlSeconds int64, allowedModelIDs []string) (token string, expiresAtUnix int64, err error)
}
