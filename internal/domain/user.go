package domain

import "time"

// User is the local LLM Studio user record.
//
// The primary key is the identra uid.
type User struct {
	ID        string
	Nickname  string
	Role      Role
	CreatedAt time.Time
	UpdatedAt time.Time
}
