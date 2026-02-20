package service

import (
	"context"
	"time"

	"github.com/ldchengyi/linkflow/internal/model"
)

// UserRepo 用户数据访问接口
type UserRepo interface {
	Create(ctx context.Context, email, passwordHash string, role model.UserRole) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByID(ctx context.Context, id string) (*model.User, error)
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
}

// TokenStore token 存储接口（Redis 操作子集）
type TokenStore interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Exists(ctx context.Context, key string) (bool, error)
	Del(ctx context.Context, keys ...string) error
	Keys(ctx context.Context, pattern string) ([]string, error)
}

// TokenGenerator token 生成/验证/撤销接口
type TokenGenerator interface {
	Generate(ctx context.Context, user *model.User) (string, error)
	Validate(ctx context.Context, tokenString string) (*Claims, error)
	Revoke(ctx context.Context, userID, token string) error
}
