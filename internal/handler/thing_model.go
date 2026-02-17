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

type ThingModelHandler struct {
	repo *repository.ThingModelRepository
	pool *pgxpool.Pool
}

func NewThingModelHandler(repo *repository.ThingModelRepository, pool *pgxpool.Pool) *ThingModelHandler {
	return &ThingModelHandler{repo: repo, pool: pool}
}

// withRLS 设置 RLS 上下文
func (h *ThingModelHandler) withRLS(c *gin.Context) (context.Context, error) {
	userID := c.GetString("user_id")
	var userRole string
	if role, exists := c.Get("user_role"); exists {
		userRole = string(role.(model.UserRole))
	}
	return database.WithRLS(c.Request.Context(), h.pool, userID, userRole)
}

// Create 创建物模型
func (h *ThingModelHandler) Create(c *gin.Context) {
	var req model.CreateThingModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	userID := c.GetString("user_id")
	var userRole string
	if role, exists := c.Get("user_role"); exists {
		userRole = string(role.(model.UserRole))
	}
	logger.Log.Infof("Creating thing model for user: %s, role: %s", userID, userRole)

	ctx, err := h.withRLS(c)
	if err != nil {
		logger.Log.Errorf("Failed to set RLS context: %v", err)
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	thingModel, err := h.repo.Create(ctx, userID, &req)
	if err != nil {
		logger.Log.Errorf("Failed to create thing model: %v", err)
		Fail(c, http.StatusInternalServerError, "failed to create thing model")
		return
	}

	Created(c, thingModel)
}

// Get 获取物模型详情
func (h *ThingModelHandler) Get(c *gin.Context) {
	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	id := c.Param("id")
	thingModel, err := h.repo.GetByID(ctx, id)
	if err != nil {
		Fail(c, http.StatusNotFound, "thing model not found")
		return
	}

	Success(c, thingModel)
}

// List 获取物模型列表
func (h *ThingModelHandler) List(c *gin.Context) {
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

	offset := (page - 1) * pageSize
	models, total, err := h.repo.List(ctx, offset, pageSize)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "failed to list thing models")
		return
	}

	Page(c, models, total, page, pageSize)
}

// Update 更新物模型
func (h *ThingModelHandler) Update(c *gin.Context) {
	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	id := c.Param("id")
	var req model.UpdateThingModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	thingModel, err := h.repo.Update(ctx, id, &req)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "failed to update thing model")
		return
	}

	Success(c, thingModel)
}

// Delete 删除物模型
func (h *ThingModelHandler) Delete(c *gin.Context) {
	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	id := c.Param("id")
	if err := h.repo.Delete(ctx, id); err != nil {
		Fail(c, http.StatusInternalServerError, "failed to delete thing model")
		return
	}

	Success(c, nil)
}
