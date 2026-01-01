package domain

import "strings"

type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleAdmin      Role = "admin"
	RoleUser       Role = "user"
)

func ParseRole(s string) (Role, bool) {
	switch Role(strings.ToLower(strings.TrimSpace(s))) {
	case RoleSuperAdmin:
		return RoleSuperAdmin, true
	case RoleAdmin:
		return RoleAdmin, true
	case RoleUser:
		return RoleUser, true
	default:
		return "", false
	}
}
