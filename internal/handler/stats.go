package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow/internal/cache"
	"github.com/ldchengyi/linkflow/internal/database"
	"github.com/ldchengyi/linkflow/internal/logger"
	"github.com/ldchengyi/linkflow/internal/model"
	"github.com/ldchengyi/linkflow/internal/repository"
)

type StatsHandler struct {
	deviceRepo     *repository.DeviceRepository
	thingModelRepo *repository.ThingModelRepository
	alertLogRepo   *repository.AlertLogRepository
	pool           *pgxpool.Pool
	rdb            *cache.Redis
}

func NewStatsHandler(
	deviceRepo *repository.DeviceRepository,
	thingModelRepo *repository.ThingModelRepository,
	alertLogRepo *repository.AlertLogRepository,
	pool *pgxpool.Pool,
	rdb *cache.Redis,
) *StatsHandler {
	return &StatsHandler{
		deviceRepo:     deviceRepo,
		thingModelRepo: thingModelRepo,
		alertLogRepo:   alertLogRepo,
		pool:           pool,
		rdb:            rdb,
	}
}

type StatsOverview struct {
	TotalDevices     int   `json:"total_devices"`
	OnlineDevices    int64 `json:"online_devices"`
	TotalThingModels int   `json:"total_thing_models"`
	TodayAlerts      int   `json:"today_alerts"`
}

// Overview 仪表盘统计概览
func (h *StatsHandler) Overview(c *gin.Context) {
	userID := c.GetString("user_id")
	var userRole string
	if role, exists := c.Get("user_role"); exists {
		userRole = string(role.(model.UserRole))
	}

	ctx, err := database.WithRLS(c.Request.Context(), h.pool, userID, userRole)
	if err != nil {
		logger.Log.Errorf("Failed to set RLS context: %v", err)
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	totalDevices, err := h.deviceRepo.Count(ctx)
	if err != nil {
		logger.Log.Errorf("Failed to count devices: %v", err)
		Fail(c, http.StatusInternalServerError, "failed to get stats")
		return
	}

	totalThingModels, err := h.thingModelRepo.Count(ctx)
	if err != nil {
		logger.Log.Errorf("Failed to count thing models: %v", err)
		Fail(c, http.StatusInternalServerError, "failed to get stats")
		return
	}

	onlineDevices, err := h.rdb.CountUserOnlineDevices(c.Request.Context(), userID)
	if err != nil {
		logger.Log.Errorf("Failed to count online devices: %v", err)
		onlineDevices = 0
	}

	todayAlerts, err := h.alertLogRepo.CountToday(ctx)
	if err != nil {
		logger.Log.Errorf("Failed to count today alerts: %v", err)
		todayAlerts = 0
	}

	Success(c, StatsOverview{
		TotalDevices:     totalDevices,
		OnlineDevices:    onlineDevices,
		TotalThingModels: totalThingModels,
		TodayAlerts:      todayAlerts,
	})
}
