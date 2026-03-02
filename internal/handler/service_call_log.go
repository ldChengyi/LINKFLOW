package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow/internal/database"
	"github.com/ldchengyi/linkflow/internal/model"
	"github.com/ldchengyi/linkflow/internal/repository"
)

type ServiceCallLogHandler struct {
	repo *repository.ServiceCallLogRepository
	pool *pgxpool.Pool
}

func NewServiceCallLogHandler(repo *repository.ServiceCallLogRepository, pool *pgxpool.Pool) *ServiceCallLogHandler {
	return &ServiceCallLogHandler{repo: repo, pool: pool}
}

func (h *ServiceCallLogHandler) withRLS(c *gin.Context) (context.Context, error) {
	userID := c.GetString("user_id")
	var userRole string
	if role, exists := c.Get("user_role"); exists {
		userRole = string(role.(model.UserRole))
	}
	return database.WithRLS(c.Request.Context(), h.pool, userID, userRole)
}

func (h *ServiceCallLogHandler) List(c *gin.Context) {
	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := h.repo.List(ctx, (page-1)*pageSize, pageSize, c.Query("device_id"))
	if err != nil {
		Fail(c, http.StatusInternalServerError, "failed to list service call logs")
		return
	}
	Page(c, logs, total, page, pageSize)
}
