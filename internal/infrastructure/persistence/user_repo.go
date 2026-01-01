package persistence

import (
	"context"
	"encoding/json"
	"time"

	"github.com/poly-workshop/llm-studio/internal/domain"
	"gorm.io/gorm"
)

// UserModel is the GORM model for LLM Studio users.
// Primary key is identra uid.
type UserModel struct {
	ID        string `gorm:"primaryKey;size:128"`
	Role      string `gorm:"size:32;index"`
	Nickname  string `gorm:"size:64"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// LoginInfoModel stores identra login info snapshot for a user.
// Primary key is identra uid (same as UserModel.ID).
type LoginInfoModel struct {
	UserID               string  `gorm:"primaryKey;size:128"`
	Email                string  `gorm:"size:320"`
	GithubID             *string `gorm:"size:128"`
	PasswordEnabled      bool
	OAuthConnectionsJSON string `gorm:"type:text"`
	UpdatedAt            time.Time
	CreatedAt            time.Time
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) EnsureExists(ctx context.Context, userID string) error {
	u := UserModel{ID: userID, Role: string(domain.RoleUser)}
	return r.db.WithContext(ctx).FirstOrCreate(&u).Error
}

func (r *UserRepository) SaveLoginInfo(ctx context.Context, info domain.LoginInfo) error {
	b, err := json.Marshal(info.OAuthConnections)
	if err != nil {
		return err
	}

	m := LoginInfoModel{
		UserID:               info.UserID,
		Email:                info.Email,
		GithubID:             info.GithubID,
		PasswordEnabled:      info.PasswordEnabled,
		OAuthConnectionsJSON: string(b),
	}
	// Upsert by primary key (works across sqlite/postgres/mysql supported by gorm).
	return r.db.WithContext(ctx).Save(&m).Error
}

func (r *UserRepository) GetRole(ctx context.Context, userID string) (domain.Role, error) {
	var u UserModel
	if err := r.db.WithContext(ctx).Select("role").First(&u, "id = ?", userID).Error; err != nil {
		return "", err
	}
	role, ok := domain.ParseRole(u.Role)
	if !ok {
		return domain.RoleUser, nil
	}
	return role, nil
}

func (r *UserRepository) SetRole(ctx context.Context, userID string, role domain.Role) error {
	// Ensure the user exists before role update.
	if err := r.EnsureExists(ctx, userID); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", userID).Update("role", string(role)).Error
}

type meRow struct {
	ID       string  `gorm:"column:id"`
	Role     string  `gorm:"column:role"`
	Nickname string  `gorm:"column:nickname"`
	Email    string  `gorm:"column:email"`
	GithubID *string `gorm:"column:github_id"`
}

func (r *UserRepository) GetMe(ctx context.Context, userID string) (domain.Me, error) {
	var row meRow
	err := r.db.WithContext(ctx).
		Model(&UserModel{}).
		Select(
			"user_models.id as id",
			"user_models.role as role",
			"user_models.nickname as nickname",
			"login_info_models.email as email",
			"login_info_models.github_id as github_id",
		).
		Joins("LEFT JOIN login_info_models ON login_info_models.user_id = user_models.id").
		Where("user_models.id = ?", userID).
		Scan(&row).Error
	if err != nil {
		return domain.Me{}, err
	}
	role, ok := domain.ParseRole(row.Role)
	if !ok {
		role = domain.RoleUser
	}
	return domain.Me{
		UserID:   row.ID,
		Role:     role,
		Email:    row.Email,
		GithubID: row.GithubID,
		Nickname: row.Nickname,
	}, nil
}

func (r *UserRepository) SetNickname(ctx context.Context, userID string, nickname string) error {
	if err := r.EnsureExists(ctx, userID); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", userID).Update("nickname", nickname).Error
}

type adminUserRow struct {
	ID              string    `gorm:"column:id"`
	Role            string    `gorm:"column:role"`
	Email           string    `gorm:"column:email"`
	GithubID        *string   `gorm:"column:github_id"`
	PasswordEnabled bool      `gorm:"column:password_enabled"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
}

func (r *UserRepository) ListUsers(ctx context.Context, limit int, offset int) ([]domain.AdminUser, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var rows []adminUserRow
	err := r.db.WithContext(ctx).
		Model(&UserModel{}).
		Select(
			"user_models.id as id",
			"user_models.role as role",
			"user_models.created_at as created_at",
			"user_models.updated_at as updated_at",
			"login_info_models.email as email",
			"login_info_models.github_id as github_id",
			"login_info_models.password_enabled as password_enabled",
		).
		Joins("LEFT JOIN login_info_models ON login_info_models.user_id = user_models.id").
		Order("user_models.created_at desc").
		Limit(limit).
		Offset(offset).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make([]domain.AdminUser, 0, len(rows))
	for _, r := range rows {
		role, ok := domain.ParseRole(r.Role)
		if !ok {
			role = domain.RoleUser
		}
		out = append(out, domain.AdminUser{
			ID:              r.ID,
			Role:            role,
			Email:           r.Email,
			GithubID:        r.GithubID,
			PasswordEnabled: r.PasswordEnabled,
			CreatedAt:       r.CreatedAt,
			UpdatedAt:       r.UpdatedAt,
		})
	}
	return out, nil
}
