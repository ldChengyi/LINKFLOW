package tts

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// EdgeTTSService 使用 edge-tts 命令行工具
type EdgeTTSService struct {
	storageDir string
}

// NewEdgeTTSService 创建服务
func NewEdgeTTSService(storageDir string) *EdgeTTSService {
	os.MkdirAll(storageDir, 0755)
	return &EdgeTTSService{storageDir: storageDir}
}

// Synthesize 生成语音，返回文件名
func (s *EdgeTTSService) Synthesize(ctx context.Context, text string) (string, error) {
	hash := md5.Sum([]byte(text))
	filename := hex.EncodeToString(hash[:]) + ".mp3"
	fullPath := filepath.Join(s.storageDir, filename)

	if _, err := os.Stat(fullPath); err == nil {
		return filename, nil
	}

	cmd := exec.CommandContext(ctx, "edge-tts", "--text", text, "--write-media", fullPath, "--voice", "zh-CN-XiaoxiaoNeural")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("edge-tts failed: %w", err)
	}

	return filename, nil
}
