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

type AlertLogHandler struct {
	repo *repository.AlertLogRepository
	pool *pgxpool.Pool
}

func NewAlertLogHandler(repo *repository.AlertLogRepository, pool *pgxpool.Pool) *AlertLogHandler {
	return &AlertLogHandler{repo: repo, pool: pool}
}

func (h *AlertLogHandler) withRLS(c *gin.Context) (context.Context, error) {
	userID := c.GetString("user_id")
	var userRole string
	if role, exists := c.Get("user_role"); exists {
		userRole = string(role.(model.UserRole))
	}
	return database.WithRLS(c.Request.Context(), h.pool, userID, userRole)
}

func (h *AlertLogHandler) List(c *gin.Context) {
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
		Fail(c, http.StatusInternalServerError, "failed to list alert logs")
		return
	}
	Page(c, logs, total, page, pageSize)
}

// UnreadCount 获取未确认告警数
func (h *AlertLogHandler) UnreadCount(c *gin.Context) {
	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	count, err := h.repo.CountUnacknowledged(ctx)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "failed to count unread alerts")
		return
	}
	Success(c, gin.H{"count": count})
}

// Acknowledge 确认告警
func (h *AlertLogHandler) Acknowledge(c *gin.Context) {
	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.repo.Acknowledge(ctx, id); err != nil {
		Fail(c, http.StatusInternalServerError, "failed to acknowledge alert")
		return
	}
	Success(c, nil)
}
