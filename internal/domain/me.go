package domain

// Me is the current-user view returned by /api/me.
type Me struct {
	UserID   string
	Role     Role
	Email    string
	GithubID *string
	Nickname string
}
