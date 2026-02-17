package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/ldchengyi/linkflow/internal/model"
	"github.com/ldchengyi/linkflow/internal/repository"
	"github.com/ldchengyi/linkflow/internal/service"
	"github.com/ldchengyi/linkflow/internal/service/mock"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	acc := testCfg.Accounts["valid_user"]
	mockRepo := mock.NewMockUserRepo(ctrl)
	mockJWT := mock.NewMockTokenGenerator(ctrl)

	mockRepo.EXPECT().Create(gomock.Any(), acc.Email, gomock.Any(), model.RoleUser).
		DoAndReturn(func(_ context.Context, email, hash string, role model.UserRole) (*model.User, error) {
			return &model.User{ID: "u1", Email: email, Role: role}, nil
		})
	mockJWT.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("tok123", nil)

	svc := service.NewAuthService(mockRepo, mockJWT)
	user, token, err := svc.Register(context.Background(), acc.Email, acc.Password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != acc.Email {
		t.Errorf("email = %s, want %s", user.Email, acc.Email)
	}
	if token != "tok123" {
		t.Errorf("token = %s, want tok123", token)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	acc := testCfg.Accounts["duplicate_user"]
	mockRepo := mock.NewMockUserRepo(ctrl)
	mockJWT := mock.NewMockTokenGenerator(ctrl)

	mockRepo.EXPECT().Create(gomock.Any(), acc.Email, gomock.Any(), model.RoleUser).
		Return(nil, repository.ErrEmailAlreadyExists)

	svc := service.NewAuthService(mockRepo, mockJWT)
	_, _, err := svc.Register(context.Background(), acc.Email, acc.Password)
	if !errors.Is(err, service.ErrEmailExists) {
		t.Fatalf("err = %v, want ErrEmailExists", err)
	}
}

func TestRegister_TokenGenError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	acc := testCfg.Accounts["valid_user"]
	mockRepo := mock.NewMockUserRepo(ctrl)
	mockJWT := mock.NewMockTokenGenerator(ctrl)

	mockRepo.EXPECT().Create(gomock.Any(), acc.Email, gomock.Any(), model.RoleUser).
		Return(&model.User{ID: "u1", Email: acc.Email}, nil)
	mockJWT.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("", errors.New("redis down"))

	svc := service.NewAuthService(mockRepo, mockJWT)
	_, _, err := svc.Register(context.Background(), acc.Email, acc.Password)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLogin_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	acc := testCfg.Accounts["login_user"]
	hash, _ := bcrypt.GenerateFromPassword([]byte(acc.Password), bcrypt.MinCost)
	mockRepo := mock.NewMockUserRepo(ctrl)
	mockJWT := mock.NewMockTokenGenerator(ctrl)

	mockRepo.EXPECT().GetByEmail(gomock.Any(), acc.Email).
		Return(&model.User{ID: "u2", Email: acc.Email, PasswordHash: string(hash)}, nil)
	mockJWT.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("tok456", nil)

	svc := service.NewAuthService(mockRepo, mockJWT)
	user, token, err := svc.Login(context.Background(), acc.Email, acc.Password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "u2" || token != "tok456" {
		t.Errorf("got user.ID=%s token=%s", user.ID, token)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	acc := testCfg.InvalidCases["nonexistent_user"]
	mockRepo := mock.NewMockUserRepo(ctrl)
	mockJWT := mock.NewMockTokenGenerator(ctrl)

	mockRepo.EXPECT().GetByEmail(gomock.Any(), acc.Email).Return(nil, repository.ErrUserNotFound)

	svc := service.NewAuthService(mockRepo, mockJWT)
	_, _, err := svc.Login(context.Background(), acc.Email, acc.Password)
	if !errors.Is(err, service.ErrInvalidCredentials) {
		t.Fatalf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	acc := testCfg.InvalidCases["wrong_password"]
	rightHash, _ := bcrypt.GenerateFromPassword([]byte("Login@12345"), bcrypt.MinCost)
	mockRepo := mock.NewMockUserRepo(ctrl)
	mockJWT := mock.NewMockTokenGenerator(ctrl)

	mockRepo.EXPECT().GetByEmail(gomock.Any(), acc.Email).
		Return(&model.User{ID: "u2", Email: acc.Email, PasswordHash: string(rightHash)}, nil)

	svc := service.NewAuthService(mockRepo, mockJWT)
	_, _, err := svc.Login(context.Background(), acc.Email, acc.Password)
	if !errors.Is(err, service.ErrInvalidCredentials) {
		t.Fatalf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestLogout_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockUserRepo(ctrl)
	mockJWT := mock.NewMockTokenGenerator(ctrl)
	mockJWT.EXPECT().Revoke(gomock.Any(), "u1", "tok").Return(nil)

	svc := service.NewAuthService(mockRepo, mockJWT)
	if err := svc.Logout(context.Background(), "u1", "tok"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
