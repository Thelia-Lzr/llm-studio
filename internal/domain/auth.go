package domain

type Token struct {
	Value     string
	ExpiresAt int64
}

type TokenPair struct {
	AccessToken  Token
	RefreshToken Token
	TokenType    string
}

// Session is the server-side session payload stored in Redis.
// UserID is the identra uid (and also the local User primary key).
type Session struct {
	UserID string
	Token  TokenPair
}
