package tests

import (
	"context"
	"testing"

	"github.com/ldchengyi/linkflow/internal/model"
	"github.com/ldchengyi/linkflow/internal/service"
	"github.com/ldchengyi/linkflow/internal/service/mock"
	"go.uber.org/mock/gomock"
)

func TestGenerate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockTokenStore(ctrl)
	store.EXPECT().Set(gomock.Any(), gomock.Any(), "1", gomock.Any()).Return(nil)

	svc := service.NewJWTService("test-secret", 1, store)
	token, err := svc.Generate(context.Background(), &model.User{
		ID: "u1", Email: "a@b.com", Role: model.RoleUser,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("token is empty")
	}
}

func TestValidate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockTokenStore(ctrl)
	store.EXPECT().Set(gomock.Any(), gomock.Any(), "1", gomock.Any()).Return(nil)
	store.EXPECT().Exists(gomock.Any(), gomock.Any()).Return(true, nil)

	svc := service.NewJWTService("test-secret", 1, store)
	user := &model.User{ID: "u1", Email: "a@b.com", Role: model.RoleUser}

	token, _ := svc.Generate(context.Background(), user)
	claims, err := svc.Validate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.UserID != "u1" || claims.Email != "a@b.com" {
		t.Errorf("claims mismatch: %+v", claims)
	}
}

func TestValidate_RevokedToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockTokenStore(ctrl)
	store.EXPECT().Set(gomock.Any(), gomock.Any(), "1", gomock.Any()).Return(nil)
	store.EXPECT().Exists(gomock.Any(), gomock.Any()).Return(false, nil)

	svc := service.NewJWTService("test-secret", 1, store)
	token, _ := svc.Generate(context.Background(), &model.User{
		ID: "u1", Email: "a@b.com", Role: model.RoleUser,
	})

	_, err := svc.Validate(context.Background(), token)
	if err == nil || err.Error() != "token revoked" {
		t.Fatalf("err = %v, want token revoked", err)
	}
}

func TestRevoke_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockTokenStore(ctrl)
	store.EXPECT().Del(gomock.Any(), gomock.Any()).Return(nil)

	svc := service.NewJWTService("test-secret", 1, store)
	if err := svc.Revoke(context.Background(), "u1", "some-token-string"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
