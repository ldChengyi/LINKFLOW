package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ldchengyi/linkflow/internal/middleware"
	"github.com/ldchengyi/linkflow/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	user, token, err := h.authService.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case service.ErrEmailExists:
			Fail(c, http.StatusConflict, "email already registered")
		default:
			Fail(c, http.StatusInternalServerError, "registration failed")
		}
		return
	}

	Created(c, AuthResponse{Token: token, User: user})
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	user, token, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			Fail(c, http.StatusUnauthorized, "invalid email or password")
		default:
			Fail(c, http.StatusInternalServerError, "login failed")
		}
		return
	}

	Success(c, AuthResponse{Token: token, User: user})
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	userID := middleware.GetUserID(c)
	token := middleware.GetToken(c)

	if err := h.authService.Logout(c.Request.Context(), userID, token); err != nil {
		Fail(c, http.StatusInternalServerError, "logout failed")
		return
	}

	Success(c, nil)
}
