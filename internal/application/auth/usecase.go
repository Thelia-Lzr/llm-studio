package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/poly-workshop/llm-studio/internal/domain"
)

type Usecase struct {
	gateway  OAuthGateway
	sessions SessionStore
	uids     UIDExtractor
	users    UserRepository

	superAdminEmailSet map[string]struct{}
}

func New(gateway OAuthGateway, sessions SessionStore, uids UIDExtractor, users UserRepository, superAdminEmails []string) *Usecase {
	set := make(map[string]struct{}, len(superAdminEmails))
	for _, e := range superAdminEmails {
		e = strings.ToLower(strings.TrimSpace(e))
		if e == "" {
			continue
		}
		set[e] = struct{}{}
	}
	return &Usecase{
		gateway:            gateway,
		sessions:           sessions,
		uids:               uids,
		users:              users,
		superAdminEmailSet: set,
	}
}

func (u *Usecase) StartOAuth(ctx context.Context, provider string, redirectURL string) (url string, state string, err error) {
	return u.gateway.GetAuthorizationURL(ctx, provider, redirectURL)
}

func (u *Usecase) CompleteOAuthToSession(ctx context.Context, code string, state string, ttl time.Duration) (string, string, error) {
	tokens, err := u.gateway.LoginByOAuth(ctx, code, state)
	if err != nil {
		return "", "", err
	}

	// Prefer identra GetCurrentUserLoginInfo (authoritative user_id + login info).
	var (
		userID    string
		loginInfo domain.LoginInfo
	)
	if li, err := u.gateway.GetCurrentUserLoginInfo(ctx, tokens.AccessToken.Value); err == nil && li.UserID != "" {
		loginInfo = li
		userID = li.UserID
	} else {
		// Fallback to extracting uid from JWT claims (best-effort).
		if u.uids == nil {
			return "", "", errors.New("uid extractor is not configured")
		}
		uid, err := u.uids.ExtractUserID(tokens.AccessToken.Value)
		if err != nil {
			return "", "", err
		}
		if uid == "" {
			return "", "", errors.New("empty user id")
		}
		userID = uid
	}

	if u.users == nil {
		return "", "", errors.New("user repository is not configured")
	}
	if err := u.users.EnsureExists(ctx, userID); err != nil {
		return "", "", err
	}
	if loginInfo.UserID != "" {
		if err := u.users.SaveLoginInfo(ctx, loginInfo); err != nil {
			return "", "", err
		}
		if u.isSuperAdminByLoginInfo(loginInfo) {
			if err := u.users.SetRole(ctx, userID, domain.RoleSuperAdmin); err != nil {
				return "", "", err
			}
		}
	}

	sessionID, err := newSessionID()
	if err != nil {
		return "", "", err
	}
	if u.sessions == nil {
		return "", "", errors.New("session store is not configured")
	}
	if err := u.sessions.Save(ctx, sessionID, domain.Session{UserID: userID, Token: tokens}, ttl); err != nil {
		return "", "", err
	}
	return sessionID, userID, nil
}

func (u *Usecase) isSuperAdminByLoginInfo(info domain.LoginInfo) bool {
	if len(u.superAdminEmailSet) == 0 {
		return false
	}
	email := strings.ToLower(strings.TrimSpace(info.Email))
	if email == "" {
		return false
	}
	_, ok := u.superAdminEmailSet[email]
	return ok
}

func (u *Usecase) Logout(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	return u.sessions.Delete(ctx, sessionID)
}

func newSessionID() (string, error) {
	// 32 bytes => 43 chars base64url (no padding). Good enough for session IDs.
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
