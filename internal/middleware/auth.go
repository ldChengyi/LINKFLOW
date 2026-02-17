package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ldchengyi/linkflow/internal/model"
	"github.com/ldchengyi/linkflow/internal/service"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	ContextUserID       = "user_id"
	ContextUserEmail    = "user_email"
	ContextUserRole     = "user_role"
	ContextToken        = "token"
)

// apiResponse 统一响应格式（middleware 包内使用）
type apiResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Auth(jwtService *service.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader(AuthorizationHeader)
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, apiResponse{Code: 401, Msg: "missing authorization header"})
			return
		}

		if !strings.HasPrefix(header, BearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, apiResponse{Code: 401, Msg: "invalid authorization format"})
			return
		}

		token := strings.TrimPrefix(header, BearerPrefix)
		claims, err := jwtService.Validate(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, apiResponse{Code: 401, Msg: err.Error()})
			return
		}

		// 将用户信息存入 context
		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUserEmail, claims.Email)
		c.Set(ContextUserRole, claims.Role)
		c.Set(ContextToken, token)

		c.Next()
	}
}

// RequireRole 角色验证中间件
func RequireRole(roles ...model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get(ContextUserRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, apiResponse{Code: 401, Msg: "unauthorized"})
			return
		}

		role := userRole.(model.UserRole)
		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, apiResponse{Code: 403, Msg: "insufficient permissions"})
	}
}

// GetUserID 从 context 获取用户ID
func GetUserID(c *gin.Context) string {
	if id, exists := c.Get(ContextUserID); exists {
		return id.(string)
	}
	return ""
}

// GetUserRole 从 context 获取用户角色
func GetUserRole(c *gin.Context) model.UserRole {
	if role, exists := c.Get(ContextUserRole); exists {
		return role.(model.UserRole)
	}
	return ""
}

// GetToken 从 context 获取 token
func GetToken(c *gin.Context) string {
	if token, exists := c.Get(ContextToken); exists {
		return token.(string)
	}
	return ""
}
