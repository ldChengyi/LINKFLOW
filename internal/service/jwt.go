package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ldchengyi/linkflow/internal/cache"
	"github.com/ldchengyi/linkflow/internal/model"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
	ErrTokenRevoked = errors.New("token revoked")
)

const tokenKeyPrefix = "token:"

// Claims JWT 声明
type Claims struct {
	UserID string         `json:"user_id"`
	Email  string         `json:"email"`
	Role   model.UserRole `json:"role"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secret      []byte
	expireHours int
	redis       *cache.Redis
}

func NewJWTService(secret string, expireHours int, redis *cache.Redis) *JWTService {
	return &JWTService{
		secret:      []byte(secret),
		expireHours: expireHours,
		redis:       redis,
	}
}

// Generate 生成 JWT token 并存入 Redis
func (s *JWTService) Generate(ctx context.Context, user *model.User) (string, error) {
	now := time.Now()
	expiration := time.Duration(s.expireHours) * time.Hour

	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "linkflow",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", err
	}

	// 存入 Redis，key 为 token:userID:tokenHash
	key := s.tokenKey(user.ID, tokenString)
	if err := s.redis.Set(ctx, key, "1", expiration); err != nil {
		return "", fmt.Errorf("store token: %w", err)
	}

	return tokenString, nil
}

// Validate 验证 JWT token（检查签名 + Redis 存在性）
func (s *JWTService) Validate(ctx context.Context, tokenString string) (*Claims, error) {
	// 1. 验证签名
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// 2. 检查 Redis 中是否存在（未被撤销）
	key := s.tokenKey(claims.UserID, tokenString)
	exists, err := s.redis.Exists(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("check token: %w", err)
	}
	if !exists {
		return nil, ErrTokenRevoked
	}

	return claims, nil
}

// Revoke 撤销 token（登出）
func (s *JWTService) Revoke(ctx context.Context, userID, tokenString string) error {
	key := s.tokenKey(userID, tokenString)
	return s.redis.Del(ctx, key)
}

// RevokeAll 撤销用户所有 token（强制登出所有设备）
func (s *JWTService) RevokeAll(ctx context.Context, userID string) error {
	pattern := fmt.Sprintf("%s%s:*", tokenKeyPrefix, userID)
	keys, err := s.redis.Client().Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return s.redis.Del(ctx, keys...)
	}
	return nil
}

func (s *JWTService) tokenKey(userID, token string) string {
	// 使用 token 前16字符作为标识（避免 key 过长）
	tokenID := token
	if len(token) > 16 {
		tokenID = token[len(token)-16:]
	}
	return fmt.Sprintf("%s%s:%s", tokenKeyPrefix, userID, tokenID)
}
