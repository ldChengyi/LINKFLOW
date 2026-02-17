package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow/internal/database"
	"github.com/ldchengyi/linkflow/internal/logger"
	"github.com/ldchengyi/linkflow/internal/model"
	"github.com/ldchengyi/linkflow/internal/repository"
)

type ScheduledTaskHandler struct {
	repo *repository.ScheduledTaskRepository
	pool *pgxpool.Pool
}

func NewScheduledTaskHandler(repo *repository.ScheduledTaskRepository, pool *pgxpool.Pool) *ScheduledTaskHandler {
	return &ScheduledTaskHandler{repo: repo, pool: pool}
}

func (h *ScheduledTaskHandler) withRLS(c *gin.Context) (context.Context, error) {
	userID := c.GetString("user_id")
	var userRole string
	if role, exists := c.Get("user_role"); exists {
		userRole = string(role.(model.UserRole))
	}
	return database.WithRLS(c.Request.Context(), h.pool, userID, userRole)
}

func (h *ScheduledTaskHandler) Create(c *gin.Context) {
	var req model.CreateScheduledTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	userID := c.GetString("user_id")
	ctx, err := h.withRLS(c)
	if err != nil {
		logger.Log.Errorf("Failed to set RLS context: %v", err)
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	task, err := h.repo.Create(ctx, userID, &req)
	if err != nil {
		logger.Log.Errorf("Failed to create scheduled task: %v", err)
		Fail(c, http.StatusInternalServerError, "failed to create scheduled task")
		return
	}
	Created(c, task)
}

func (h *ScheduledTaskHandler) Get(c *gin.Context) {
	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	task, err := h.repo.GetByID(ctx, c.Param("id"))
	if err != nil {
		Fail(c, http.StatusNotFound, "scheduled task not found")
		return
	}
	Success(c, task)
}

func (h *ScheduledTaskHandler) List(c *gin.Context) {
	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	tasks, total, err := h.repo.List(ctx, (page-1)*pageSize, pageSize, c.Query("device_id"))
	if err != nil {
		Fail(c, http.StatusInternalServerError, "failed to list scheduled tasks")
		return
	}
	Page(c, tasks, total, page, pageSize)
}

func (h *ScheduledTaskHandler) Update(c *gin.Context) {
	var req model.UpdateScheduledTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	task, err := h.repo.Update(ctx, c.Param("id"), &req)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "failed to update scheduled task")
		return
	}
	Success(c, task)
}

func (h *ScheduledTaskHandler) Delete(c *gin.Context) {
	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	if err := h.repo.Delete(ctx, c.Param("id")); err != nil {
		Fail(c, http.StatusInternalServerError, "failed to delete scheduled task")
		return
	}
	Success(c, nil)
}
