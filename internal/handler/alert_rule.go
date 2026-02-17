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

// AlertRuleBrokerInvalidator 用于通知 Broker 清除告警规则缓存
type AlertRuleBrokerInvalidator interface {
	InvalidateAlertRulesCache(deviceID string)
}

type AlertRuleHandler struct {
	repo        *repository.AlertRuleRepository
	pool        *pgxpool.Pool
	invalidator AlertRuleBrokerInvalidator
}

func NewAlertRuleHandler(repo *repository.AlertRuleRepository, pool *pgxpool.Pool, invalidator AlertRuleBrokerInvalidator) *AlertRuleHandler {
	return &AlertRuleHandler{repo: repo, pool: pool, invalidator: invalidator}
}

func (h *AlertRuleHandler) withRLS(c *gin.Context) (context.Context, error) {
	userID := c.GetString("user_id")
	var userRole string
	if role, exists := c.Get("user_role"); exists {
		userRole = string(role.(model.UserRole))
	}
	return database.WithRLS(c.Request.Context(), h.pool, userID, userRole)
}

func (h *AlertRuleHandler) Create(c *gin.Context) {
	var req model.CreateAlertRuleRequest
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

	rule, err := h.repo.Create(ctx, userID, &req)
	if err != nil {
		logger.Log.Errorf("Failed to create alert rule: %v", err)
		Fail(c, http.StatusInternalServerError, "failed to create alert rule")
		return
	}

	// 清除 Broker 缓存
	if h.invalidator != nil {
		h.invalidator.InvalidateAlertRulesCache(req.DeviceID)
	}

	Created(c, rule)
}

func (h *AlertRuleHandler) Get(c *gin.Context) {
	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	rule, err := h.repo.GetByID(ctx, c.Param("id"))
	if err != nil {
		Fail(c, http.StatusNotFound, "alert rule not found")
		return
	}
	Success(c, rule)
}

func (h *AlertRuleHandler) List(c *gin.Context) {
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

	rules, total, err := h.repo.List(ctx, (page-1)*pageSize, pageSize, c.Query("device_id"))
	if err != nil {
		Fail(c, http.StatusInternalServerError, "failed to list alert rules")
		return
	}
	Page(c, rules, total, page, pageSize)
}

func (h *AlertRuleHandler) Update(c *gin.Context) {
	var req model.UpdateAlertRuleRequest
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

	// 获取旧规则以清除旧设备缓存
	oldRule, _ := h.repo.GetByID(ctx, c.Param("id"))

	rule, err := h.repo.Update(ctx, c.Param("id"), &req)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "failed to update alert rule")
		return
	}

	// 清除新旧设备的 Broker 缓存
	if h.invalidator != nil {
		h.invalidator.InvalidateAlertRulesCache(req.DeviceID)
		if oldRule != nil && oldRule.DeviceID != req.DeviceID {
			h.invalidator.InvalidateAlertRulesCache(oldRule.DeviceID)
		}
	}

	Success(c, rule)
}

func (h *AlertRuleHandler) Delete(c *gin.Context) {
	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	// 获取规则以清除设备缓存
	rule, _ := h.repo.GetByID(ctx, c.Param("id"))

	if err := h.repo.Delete(ctx, c.Param("id")); err != nil {
		Fail(c, http.StatusInternalServerError, "failed to delete alert rule")
		return
	}

	if h.invalidator != nil && rule != nil {
		h.invalidator.InvalidateAlertRulesCache(rule.DeviceID)
	}

	Success(c, nil)
}
