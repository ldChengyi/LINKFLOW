package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ldchengyi/linkflow/internal/middleware"
	"github.com/ldchengyi/linkflow/internal/model"
	"github.com/ldchengyi/linkflow/internal/repository"
)

type AuditLogHandler struct {
	repo *repository.AuditLogRepository
}

func NewAuditLogHandler(repo *repository.AuditLogRepository) *AuditLogHandler {
	return &AuditLogHandler{repo: repo}
}

// List 查询审计日志
// admin 可查所有，普通用户只能查自己的
func (h *AuditLogHandler) List(c *gin.Context) {
	var q model.AuditLogQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		Fail(c, http.StatusBadRequest, "invalid query params")
		return
	}

	// 非 admin 强制只能查自己的日志
	role := middleware.GetUserRole(c)
	if role != model.RoleAdmin {
		q.UserID = middleware.GetUserID(c)
	}

	list, total, err := h.repo.List(c.Request.Context(), &q)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "query audit logs failed")
		return
	}
	if list == nil {
		list = []*model.AuditLog{}
	}

	Page(c, list, total, q.Page, q.PageSize)
}
