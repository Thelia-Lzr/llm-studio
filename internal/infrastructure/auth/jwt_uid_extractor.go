package auth

import (
	"errors"

	jwt "github.com/golang-jwt/jwt/v5"
)

// JWTUIDExtractor extracts identra uid from an access token JWT.
//
// NOTE: This implementation parses claims without signature validation.
// The assumption is that the access token originates from identra over a trusted channel.
// If you need stronger guarantees, validate signature using identra JWKS before extracting claims.
type JWTUIDExtractor struct{}

func NewJWTUIDExtractor() *JWTUIDExtractor { return &JWTUIDExtractor{} }

func (e *JWTUIDExtractor) ExtractUserID(accessToken string) (string, error) {
	if accessToken == "" {
		return "", errors.New("empty access token")
	}

	// Parse without verification (we only need claims to locate uid/sub).
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(accessToken, jwt.MapClaims{})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("unexpected jwt claims type")
	}

	// Prefer "uid" if present, fallback to JWT standard "sub".
	if v, ok := claims["uid"]; ok {
		if s, ok := v.(string); ok && s != "" {
			return s, nil
		}
	}
	if v, ok := claims["sub"]; ok {
		if s, ok := v.(string); ok && s != "" {
			return s, nil
		}
	}
	return "", errors.New("user id not found in jwt claims")
}
