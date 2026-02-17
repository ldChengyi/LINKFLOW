//go:build integration

package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/ldchengyi/linkflow/internal/repository"
	"github.com/ldchengyi/linkflow/internal/service"
)

func TestIntegration_RegisterAndLogin(t *testing.T) {
	acc := testCfg.Accounts["valid_user"]
	userRepo := repository.NewUserRepository(testPool)
	jwtSvc := service.NewJWTService(
		testCfg.Integration.JWT.Secret,
		testCfg.Integration.JWT.ExpireHours,
		testRdb,
	)
	authSvc := service.NewAuthService(userRepo, jwtSvc)
	ctx := context.Background()

	testPool.Exec(ctx, "DELETE FROM users WHERE email = $1", acc.Email)

	user, token, err := authSvc.Register(ctx, acc.Email, acc.Password)
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if user.Email != acc.Email || token == "" {
		t.Fatalf("register result: user=%+v token=%s", user, token)
	}

	user2, token2, err := authSvc.Login(ctx, acc.Email, acc.Password)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if user2.ID != user.ID || token2 == "" {
		t.Fatalf("login result mismatch")
	}

	claims, err := jwtSvc.Validate(ctx, token2)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if claims.UserID != user.ID {
		t.Errorf("claims.UserID = %s, want %s", claims.UserID, user.ID)
	}
}

func TestIntegration_DuplicateRegister(t *testing.T) {
	acc := testCfg.Accounts["duplicate_user"]
	userRepo := repository.NewUserRepository(testPool)
	jwtSvc := service.NewJWTService(
		testCfg.Integration.JWT.Secret,
		testCfg.Integration.JWT.ExpireHours,
		testRdb,
	)
	authSvc := service.NewAuthService(userRepo, jwtSvc)
	ctx := context.Background()

	testPool.Exec(ctx, "DELETE FROM users WHERE email = $1", acc.Email)

	_, _, err := authSvc.Register(ctx, acc.Email, acc.Password)
	if err != nil {
		t.Fatalf("first register: %v", err)
	}

	_, _, err = authSvc.Register(ctx, acc.Email, acc.Password)
	if !errors.Is(err, service.ErrEmailExists) {
		t.Fatalf("err = %v, want ErrEmailExists", err)
	}
}

func TestIntegration_LoginWrongPassword(t *testing.T) {
	acc := testCfg.Accounts["login_user"]
	wrong := testCfg.InvalidCases["wrong_password"]
	userRepo := repository.NewUserRepository(testPool)
	jwtSvc := service.NewJWTService(
		testCfg.Integration.JWT.Secret,
		testCfg.Integration.JWT.ExpireHours,
		testRdb,
	)
	authSvc := service.NewAuthService(userRepo, jwtSvc)
	ctx := context.Background()

	testPool.Exec(ctx, "DELETE FROM users WHERE email = $1", acc.Email)
	_, _, _ = authSvc.Register(ctx, acc.Email, acc.Password)

	_, _, err := authSvc.Login(ctx, wrong.Email, wrong.Password)
	if !errors.Is(err, service.ErrInvalidCredentials) {
		t.Fatalf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestIntegration_LogoutRevokesToken(t *testing.T) {
	acc := testCfg.Accounts["logout_user"]
	userRepo := repository.NewUserRepository(testPool)
	jwtSvc := service.NewJWTService(
		testCfg.Integration.JWT.Secret,
		testCfg.Integration.JWT.ExpireHours,
		testRdb,
	)
	authSvc := service.NewAuthService(userRepo, jwtSvc)
	ctx := context.Background()

	testPool.Exec(ctx, "DELETE FROM users WHERE email = $1", acc.Email)

	user, token, err := authSvc.Register(ctx, acc.Email, acc.Password)
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if err := authSvc.Logout(ctx, user.ID, token); err != nil {
		t.Fatalf("logout: %v", err)
	}

	_, err = jwtSvc.Validate(ctx, token)
	if err == nil {
		t.Fatal("expected token revoked error")
	}
}
