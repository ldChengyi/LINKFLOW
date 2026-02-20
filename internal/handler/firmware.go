package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow/internal/database"
	"github.com/ldchengyi/linkflow/internal/logger"
	"github.com/ldchengyi/linkflow/internal/model"
	"github.com/ldchengyi/linkflow/internal/repository"
)

const firmwareUploadDir = "uploads/firmwares"

type FirmwareHandler struct {
	repo       *repository.FirmwareRepository
	deviceRepo *repository.DeviceRepository
	pool       *pgxpool.Pool
}

func NewFirmwareHandler(repo *repository.FirmwareRepository, deviceRepo *repository.DeviceRepository, pool *pgxpool.Pool) *FirmwareHandler {
	os.MkdirAll(firmwareUploadDir, 0755)
	return &FirmwareHandler{repo: repo, deviceRepo: deviceRepo, pool: pool}
}

func (h *FirmwareHandler) withRLS(c *gin.Context) (context.Context, error) {
	userID := c.GetString("user_id")
	var userRole string
	if role, exists := c.Get("user_role"); exists {
		userRole = string(role.(model.UserRole))
	}
	return database.WithRLS(c.Request.Context(), h.pool, userID, userRole)
}

// Upload 上传固件 (multipart/form-data)
func (h *FirmwareHandler) Upload(c *gin.Context) {
	name := c.PostForm("name")
	version := c.PostForm("version")
	description := c.PostForm("description")
	if name == "" || version == "" {
		Fail(c, http.StatusBadRequest, "name and version are required")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		Fail(c, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	// 计算 SHA256 并保存文件
	hasher := sha256.New()
	tmpPath := filepath.Join(firmwareUploadDir, "tmp_"+header.Filename)
	out, err := os.Create(tmpPath)
	if err != nil {
		logger.Log.Errorf("Failed to create temp file: %v", err)
		Fail(c, http.StatusInternalServerError, "failed to save file")
		return
	}

	size, err := io.Copy(io.MultiWriter(out, hasher), file)
	out.Close()
	if err != nil {
		os.Remove(tmpPath)
		Fail(c, http.StatusInternalServerError, "failed to save file")
		return
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	finalPath := filepath.Join(firmwareUploadDir, checksum+filepath.Ext(header.Filename))
	os.Rename(tmpPath, finalPath)

	userID := c.GetString("user_id")
	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	fw, err := h.repo.Create(ctx, &model.Firmware{
		UserID:      userID,
		Name:        name,
		Version:     version,
		FilePath:    finalPath,
		FileSize:    size,
		Checksum:    checksum,
		Description: description,
	})
	if err != nil {
		logger.Log.Errorf("Failed to create firmware: %v", err)
		Fail(c, http.StatusInternalServerError, "failed to create firmware")
		return
	}
	Created(c, fw)
}

func (h *FirmwareHandler) List(c *gin.Context) {
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

	list, total, err := h.repo.List(ctx, (page-1)*pageSize, pageSize)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "failed to list firmwares")
		return
	}
	Page(c, list, total, page, pageSize)
}

func (h *FirmwareHandler) Delete(c *gin.Context) {
	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	fw, err := h.repo.GetByID(ctx, c.Param("id"))
	if err != nil {
		Fail(c, http.StatusNotFound, "firmware not found")
		return
	}

	if err := h.repo.Delete(ctx, c.Param("id")); err != nil {
		Fail(c, http.StatusInternalServerError, "failed to delete firmware")
		return
	}

	os.Remove(fw.FilePath)
	Success(c, nil)
}

// Download 设备下载固件 (Basic Auth: device_id:device_secret)
func (h *FirmwareHandler) Download(c *gin.Context) {
	deviceID, secret, ok := c.Request.BasicAuth()
	if !ok {
		c.Header("WWW-Authenticate", `Basic realm="firmware"`)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if _, err := h.deviceRepo.AuthenticateDevice(c.Request.Context(), deviceID, secret); err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// 用 admin pool 直接查固件（不走 RLS）
	fw, err := h.repo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filepath.Base(fw.FilePath))
	c.Header("X-Checksum", fw.Checksum)
	c.File(fw.FilePath)
}
