package service

import (
	"context"
	"errors"

	"github.com/ldchengyi/linkflow/internal/model"
	"github.com/ldchengyi/linkflow/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("email already registered")
)

type AuthService struct {
	userRepo   UserRepo
	jwtService TokenGenerator
}

func NewAuthService(userRepo UserRepo, jwtService TokenGenerator) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

// Register 用户注册
func (s *AuthService) Register(ctx context.Context, email, password string) (*model.User, string, error) {
	// 密码哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	// 创建用户（默认角色为 user）
	user, err := s.userRepo.Create(ctx, email, string(hash), model.RoleUser)
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyExists) {
			return nil, "", ErrEmailExists
		}
		return nil, "", err
	}

	// 生成 token
	token, err := s.jwtService.Generate(ctx, user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, email, password string) (*model.User, string, error) {
	// 查询用户
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	// 生成 token
	token, err := s.jwtService.Generate(ctx, user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// Logout 用户登出（撤销 token）
func (s *AuthService) Logout(ctx context.Context, userID, token string) error {
	return s.jwtService.Revoke(ctx, userID, token)
}
