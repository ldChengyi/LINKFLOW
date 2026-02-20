package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow/internal/database"
	"github.com/ldchengyi/linkflow/internal/logger"
	"github.com/ldchengyi/linkflow/internal/model"
	"github.com/ldchengyi/linkflow/internal/repository"
)

type OTATaskHandler struct {
	taskRepo     *repository.OTATaskRepository
	firmwareRepo *repository.FirmwareRepository
	deviceRepo   *repository.DeviceRepository
	pool         *pgxpool.Pool
	publisher    MQTTPublisher
	baseURL      string
}

func NewOTATaskHandler(
	taskRepo *repository.OTATaskRepository,
	firmwareRepo *repository.FirmwareRepository,
	deviceRepo *repository.DeviceRepository,
	pool *pgxpool.Pool,
	publisher MQTTPublisher,
	baseURL string,
) *OTATaskHandler {
	return &OTATaskHandler{
		taskRepo: taskRepo, firmwareRepo: firmwareRepo,
		deviceRepo: deviceRepo, pool: pool,
		publisher: publisher, baseURL: baseURL,
	}
}

func (h *OTATaskHandler) withRLS(c *gin.Context) (context.Context, error) {
	userID := c.GetString("user_id")
	var userRole string
	if role, exists := c.Get("user_role"); exists {
		userRole = string(role.(model.UserRole))
	}
	return database.WithRLS(c.Request.Context(), h.pool, userID, userRole)
}

func (h *OTATaskHandler) Create(c *gin.Context) {
	var req model.CreateOTATaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	userID := c.GetString("user_id")
	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	// 查设备
	device, err := h.deviceRepo.GetByID(ctx, req.DeviceID)
	if err != nil {
		Fail(c, http.StatusNotFound, "device not found")
		return
	}

	// 查固件
	fw, err := h.firmwareRepo.GetByID(ctx, req.FirmwareID)
	if err != nil {
		Fail(c, http.StatusNotFound, "firmware not found")
		return
	}

	task, err := h.taskRepo.Create(ctx, &model.OTATask{
		UserID:          userID,
		DeviceID:        req.DeviceID,
		DeviceName:      device.Name,
		FirmwareID:      req.FirmwareID,
		FirmwareVersion: fw.Version,
	})
	if err != nil {
		logger.Log.Errorf("Failed to create OTA task: %v", err)
		Fail(c, http.StatusInternalServerError, "failed to create OTA task")
		return
	}

	// 设备在线则立即下发
	if h.publisher != nil && h.publisher.IsClientConnected(req.DeviceID) {
		go h.pushOTACommand(task.ID, req.DeviceID, fw)
	}

	Created(c, task)
}

func (h *OTATaskHandler) pushOTACommand(taskID, deviceID string, fw *model.Firmware) {
	downloadURL := fmt.Sprintf("%s/api/firmwares/%s/download", h.baseURL, fw.ID)
	cmd := map[string]interface{}{
		"task_id":  taskID,
		"version":  fw.Version,
		"url":      downloadURL,
		"checksum": fw.Checksum,
		"size":     fw.FileSize,
	}
	payload, _ := json.Marshal(cmd)
	topic := fmt.Sprintf("devices/%s/ota/down", deviceID)

	if err := h.publisher.Publish(topic, payload, false, 1); err != nil {
		logger.Log.Errorf("Failed to push OTA command: task=%s, err=%v", taskID, err)
		return
	}

	// 更新状态为 pushing
	ctx := context.Background()
	h.taskRepo.UpdateProgress(ctx, taskID, "pushing", 0, "")
}

func (h *OTATaskHandler) List(c *gin.Context) {
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

	list, total, err := h.taskRepo.List(ctx, (page-1)*pageSize, pageSize, c.Query("device_id"))
	if err != nil {
		Fail(c, http.StatusInternalServerError, "failed to list OTA tasks")
		return
	}
	Page(c, list, total, page, pageSize)
}

func (h *OTATaskHandler) Get(c *gin.Context) {
	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	task, err := h.taskRepo.GetByID(ctx, c.Param("id"))
	if err != nil {
		Fail(c, http.StatusNotFound, "OTA task not found")
		return
	}
	Success(c, task)
}

func (h *OTATaskHandler) Cancel(c *gin.Context) {
	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	if err := h.taskRepo.Cancel(ctx, c.Param("id")); err != nil {
		Fail(c, http.StatusInternalServerError, "failed to cancel OTA task")
		return
	}
	Success(c, nil)
}
