package domain

// OAuthConnection represents one OAuth identity linked to the user in identra.
type OAuthConnection struct {
	Provider       string
	ProviderUserID string
}

// LoginInfo is the user's login information returned by identra.
type LoginInfo struct {
	UserID           string
	Email            string
	GithubID         *string
	PasswordEnabled  bool
	OAuthConnections []OAuthConnection
}
