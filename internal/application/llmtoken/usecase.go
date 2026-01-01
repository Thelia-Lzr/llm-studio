package llmtoken

import (
	"context"
	"errors"
)

var (
	ErrUnauthenticated = errors.New("unauthenticated")
)

type Usecase struct {
	sessions        SessionStore
	issuer          TokenIssuer
	ttlSeconds      int64
	allowedModelIDs []string
}

func New(sessions SessionStore, issuer TokenIssuer, ttlSeconds int64, allowedModelIDs []string) *Usecase {
	return &Usecase{sessions: sessions, issuer: issuer, ttlSeconds: ttlSeconds, allowedModelIDs: allowedModelIDs}
}

func (u *Usecase) IssueToken(ctx context.Context, sessionID string) (token string, expiresAtUnix int64, err error) {
	if sessionID == "" {
		return "", 0, ErrUnauthenticated
	}
	if u == nil || u.sessions == nil || u.issuer == nil {
		return "", 0, errors.New("token issuer not configured")
	}
	sess, err := u.sessions.Get(ctx, sessionID)
	if err != nil {
		return "", 0, ErrUnauthenticated
	}
	return u.issuer.IssueToken(ctx, sess.UserID, u.ttlSeconds, u.allowedModelIDs)
}
