package model

import "time"

// UserRole 用户角色
type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleUser   UserRole = "user"
	RoleViewer UserRole = "viewer"
)

// User 用户模型
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         UserRole  `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
