package handler

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/ldchengyi/linkflow/internal/tts"
)

// TTSHandler TTS 音频文件下载 + 测试
type TTSHandler struct {
	storageDir string
	service    tts.Service
}

// NewTTSHandler 创建 handler
func NewTTSHandler(storageDir string, service tts.Service) *TTSHandler {
	return &TTSHandler{storageDir: storageDir, service: service}
}

// Download 下载音频文件
func (h *TTSHandler) Download(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename required"})
		return
	}

	filePath := filepath.Join(h.storageDir, filename)
	c.File(filePath)
}

// Test POST /api/tts/test — 测试 TTS 合成
func (h *TTSHandler) Test(c *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "text is required")
		return
	}

	if len([]rune(req.Text)) > 200 {
		Fail(c, http.StatusBadRequest, "text too long, max 200 characters")
		return
	}

	filename, err := h.service.Synthesize(c.Request.Context(), req.Text)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "TTS synthesis failed: "+err.Error())
		return
	}

	audioURL := "/api/tts/" + filename
	Success(c, gin.H{"audio_url": audioURL})
}
