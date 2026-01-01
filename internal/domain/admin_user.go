package domain

import "time"

// AdminUser is the user view for the user-management page.
type AdminUser struct {
	ID              string
	Role            Role
	Email           string
	GithubID        *string
	PasswordEnabled bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
