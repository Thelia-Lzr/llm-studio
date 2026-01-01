package rbac

import (
	"context"
	"errors"
	"strings"

	"github.com/poly-workshop/llm-studio/internal/domain"
)

var (
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrForbidden       = errors.New("forbidden")
	ErrInvalidRole     = errors.New("invalid role")
	ErrInvalidNickname = errors.New("invalid nickname")
)

type Usecase struct {
	sessions SessionStore
	users    UserRepository
}

func New(sessions SessionStore, users UserRepository) *Usecase {
	return &Usecase{sessions: sessions, users: users}
}

func (u *Usecase) Me(ctx context.Context, sessionID string) (domain.Me, error) {
	if sessionID == "" {
		return domain.Me{}, ErrUnauthenticated
	}
	sess, err := u.sessions.Get(ctx, sessionID)
	if err != nil {
		return domain.Me{}, ErrUnauthenticated
	}
	return u.users.GetMe(ctx, sess.UserID)
}

func (u *Usecase) UpdateMyNickname(ctx context.Context, sessionID string, nickname string) (domain.Me, error) {
	if sessionID == "" {
		return domain.Me{}, ErrUnauthenticated
	}
	sess, err := u.sessions.Get(ctx, sessionID)
	if err != nil {
		return domain.Me{}, ErrUnauthenticated
	}
	nickname = strings.TrimSpace(nickname)
	if len(nickname) > 32 {
		return domain.Me{}, ErrInvalidNickname
	}
	if err := u.users.SetNickname(ctx, sess.UserID, nickname); err != nil {
		return domain.Me{}, err
	}
	return u.users.GetMe(ctx, sess.UserID)
}

func (u *Usecase) ListUsers(ctx context.Context, sessionID string, limit int, offset int) ([]domain.AdminUser, error) {
	if sessionID == "" {
		return nil, ErrUnauthenticated
	}
	sess, err := u.sessions.Get(ctx, sessionID)
	if err != nil {
		return nil, ErrUnauthenticated
	}
	role, err := u.users.GetRole(ctx, sess.UserID)
	if err != nil {
		return nil, err
	}
	if role != domain.RoleSuperAdmin && role != domain.RoleAdmin {
		return nil, ErrForbidden
	}
	return u.users.ListUsers(ctx, limit, offset)
}

// SetUserRole allows a super admin to assign admin/user role to a target user.
// Super admin is config-driven and cannot be granted via this API.
func (u *Usecase) SetUserRole(ctx context.Context, actorSessionID string, targetUserID string, role domain.Role) error {
	if actorSessionID == "" {
		return ErrUnauthenticated
	}
	actorSess, err := u.sessions.Get(ctx, actorSessionID)
	if err != nil {
		return ErrUnauthenticated
	}
	actorRole, err := u.users.GetRole(ctx, actorSess.UserID)
	if err != nil {
		return err
	}
	if actorRole != domain.RoleSuperAdmin {
		return ErrForbidden
	}
	if role != domain.RoleAdmin && role != domain.RoleUser {
		return ErrInvalidRole
	}

	// Prevent demoting a user currently marked as super_admin in DB (config-driven).
	if cur, err := u.users.GetRole(ctx, targetUserID); err == nil && cur == domain.RoleSuperAdmin {
		return ErrForbidden
	}

	return u.users.SetRole(ctx, targetUserID, role)
}
