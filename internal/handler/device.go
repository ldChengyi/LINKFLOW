package handler

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow/internal/cache"
	"github.com/ldchengyi/linkflow/internal/database"
	"github.com/ldchengyi/linkflow/internal/logger"
	"github.com/ldchengyi/linkflow/internal/model"
	"github.com/ldchengyi/linkflow/internal/repository"
)

type DeviceHandler struct {
	repo     *repository.DeviceRepository
	dataRepo *repository.DeviceDataRepository
	pool     *pgxpool.Pool
	rdb      *cache.Redis
}

func NewDeviceHandler(repo *repository.DeviceRepository, dataRepo *repository.DeviceDataRepository, pool *pgxpool.Pool, rdb *cache.Redis) *DeviceHandler {
	return &DeviceHandler{repo: repo, dataRepo: dataRepo, pool: pool, rdb: rdb}
}

func (h *DeviceHandler) withRLS(c *gin.Context) (context.Context, error) {
	userID := c.GetString("user_id")
	var userRole string
	if role, exists := c.Get("user_role"); exists {
		userRole = string(role.(model.UserRole))
	}
	return database.WithRLS(c.Request.Context(), h.pool, userID, userRole)
}

// mergeOnlineStatus 用 Redis 实时状态覆写设备的 Status 字段
func (h *DeviceHandler) mergeOnlineStatus(ctx context.Context, devices []*model.Device) {
	ids := make([]string, len(devices))
	for i, d := range devices {
		ids[i] = d.ID
	}
	statusMap, err := h.rdb.BatchCheckOnline(ctx, ids)
	if err != nil {
		logger.Log.Errorf("Failed to batch check online status: %v", err)
		return // 降级使用 PG 状态
	}
	for _, d := range devices {
		if statusMap[d.ID] {
			d.Status = "online"
		} else {
			d.Status = "offline"
		}
	}
}

// Create 创建设备
func (h *DeviceHandler) Create(c *gin.Context) {
	var req model.CreateDeviceRequest
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

	device, err := h.repo.Create(ctx, userID, &req)
	if err != nil {
		logger.Log.Errorf("Failed to create device: %v", err)
		Fail(c, http.StatusInternalServerError, "failed to create device")
		return
	}

	Created(c, device)
}

// Get 获取设备详情
func (h *DeviceHandler) Get(c *gin.Context) {
	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	id := c.Param("id")
	device, err := h.repo.GetByID(ctx, id)
	if err != nil {
		Fail(c, http.StatusNotFound, "device not found")
		return
	}

	// 合并 Redis 实时在线状态
	h.mergeOnlineStatus(c.Request.Context(), []*model.Device{device})

	Success(c, device)
}

// List 获取设备列表
func (h *DeviceHandler) List(c *gin.Context) {
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
	devices, total, err := h.repo.List(ctx, offset, pageSize)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "failed to list devices")
		return
	}

	// 合并 Redis 实时在线状态
	h.mergeOnlineStatus(c.Request.Context(), devices)

	Page(c, devices, total, page, pageSize)
}

// Update 更新设备
func (h *DeviceHandler) Update(c *gin.Context) {
	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	id := c.Param("id")
	var req model.UpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	device, err := h.repo.Update(ctx, id, &req)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "failed to update device")
		return
	}

	Success(c, device)
}

// Delete 删除设备
func (h *DeviceHandler) Delete(c *gin.Context) {
	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	id := c.Param("id")
	if err := h.repo.Delete(ctx, id); err != nil {
		Fail(c, http.StatusInternalServerError, "failed to delete device")
		return
	}

	Success(c, nil)
}

// History 获取设备历史遥测数据（根据时间范围自动选择聚合粒度）
func (h *DeviceHandler) History(c *gin.Context) {
	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	id := c.Param("id")
	if _, err = h.repo.GetByID(ctx, id); err != nil {
		Fail(c, http.StatusNotFound, "device not found")
		return
	}

	end := time.Now()
	start := end.Add(-1 * time.Hour)
	if s := c.Query("start"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			start = t
		}
	}
	if e := c.Query("end"); e != "" {
		if t, err := time.Parse(time.RFC3339, e); err == nil {
			end = t
		}
	}

	duration := end.Sub(start)

	// 根据时长自动选择聚合粒度
	// ≤2h → 原始数据; ≤12h → 5min; ≤48h → 15min; ≤7d → 1h; >7d → 6h
	var interval string
	switch {
	case duration <= 2*time.Hour:
		interval = "" // 原始数据
	case duration <= 12*time.Hour:
		interval = "5 minutes"
	case duration <= 48*time.Hour:
		interval = "15 minutes"
	case duration <= 7*24*time.Hour:
		interval = "1 hour"
	default:
		interval = "6 hours"
	}

	if interval == "" {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "200"))
		if limit < 1 || limit > 1000 {
			limit = 200
		}
		data, err := h.dataRepo.GetDataHistory(ctx, id, start, end, limit)
		if err != nil {
			logger.Log.Errorf("Failed to get history data for device %s: %v", id, err)
			Fail(c, http.StatusInternalServerError, "failed to get history data")
			return
		}
		Success(c, gin.H{
			"aggregated": false,
			"interval":   "",
			"data":       data,
		})
		return
	}

	aggData, err := h.dataRepo.GetDataHistoryAggregated(ctx, id, start, end, interval)
	if err != nil {
		logger.Log.Errorf("Failed to get aggregated history data for device %s: %v", id, err)
		Fail(c, http.StatusInternalServerError, "failed to get history data")
		return
	}
	Success(c, gin.H{
		"aggregated": true,
		"interval":   interval,
		"data":       aggData,
	})
}

// ExportCSV 导出设备历史数据为 CSV
func (h *DeviceHandler) ExportCSV(c *gin.Context) {
	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	id := c.Param("id")
	if _, err = h.repo.GetByID(ctx, id); err != nil {
		Fail(c, http.StatusNotFound, "device not found")
		return
	}

	end := time.Now()
	start := end.Add(-1 * time.Hour)
	if s := c.Query("start"); s != "" {
		if t, err2 := time.Parse(time.RFC3339, s); err2 == nil {
			start = t
		}
	}
	if e := c.Query("end"); e != "" {
		if t, err2 := time.Parse(time.RFC3339, e); err2 == nil {
			end = t
		}
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "1000"))
	if limit < 1 || limit > 5000 {
		limit = 1000
	}

	data, err := h.dataRepo.GetDataHistory(ctx, id, start, end, limit)
	if err != nil {
		logger.Log.Errorf("ExportCSV: failed to get history data for device %s: %v", id, err)
		Fail(c, http.StatusInternalServerError, "failed to get history data")
		return
	}

	// 收集所有 payload key
	keySet := make(map[string]struct{})
	for _, d := range data {
		for k := range d.Payload {
			keySet[k] = struct{}{}
		}
	}
	keys := make([]string, 0, len(keySet))
	for k := range keySet {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 写 CSV
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	header := append([]string{"time", "valid"}, keys...)
	w.Write(header)

	for _, d := range data {
		row := make([]string, 0, len(header))
		row = append(row, d.Time.UTC().Format(time.RFC3339), fmt.Sprintf("%v", d.Valid))
		for _, k := range keys {
			if v, ok := d.Payload[k]; ok {
				row = append(row, fmt.Sprintf("%v", v))
			} else {
				row = append(row, "")
			}
		}
		w.Write(row)
	}
	w.Flush()

	filename := fmt.Sprintf("device_%s_%s.csv", id[:8], start.UTC().Format("20060102"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", buf.Bytes())
}

// LatestData 获取设备最新遥测数据
func (h *DeviceHandler) LatestData(c *gin.Context) {
	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	id := c.Param("id")

	// 先验证设备存在且属于当前用户（RLS 会过滤）
	_, err = h.repo.GetByID(ctx, id)
	if err != nil {
		Fail(c, http.StatusNotFound, "device not found")
		return
	}

	data, err := h.dataRepo.GetLatestData(ctx, id)
	if err != nil {
		logger.Log.Errorf("Failed to get latest data for device %s: %v", id, err)
		Fail(c, http.StatusInternalServerError, "failed to get latest data")
		return
	}

	Success(c, data)
}
