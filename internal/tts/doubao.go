package tts

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ldchengyi/linkflow/internal/logger"
)

const doubaoTTSURL = "https://openspeech.bytedance.com/api/v3/tts/unidirectional"

// doubaoRequest 豆包 TTS 请求体
type doubaoRequest struct {
	User      doubaoUser      `json:"user"`
	ReqParams doubaoReqParams `json:"req_params"`
}

type doubaoUser struct {
	UID string `json:"uid"`
}

type doubaoReqParams struct {
	Text        string           `json:"text"`
	Speaker     string           `json:"speaker"`
	AudioParams doubaoAudioParams `json:"audio_params"`
}

type doubaoAudioParams struct {
	Format     string `json:"format"`
	SampleRate int    `json:"sample_rate"`
}

// doubaoChunk HTTP Chunked 响应中的每个 JSON 块
type doubaoChunk struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"` // base64 编码的音频片段
}

// DoubaoTTSService 豆包声音复刻 TTS 服务
type DoubaoTTSService struct {
	storageDir string
	provider   TTSSettingsProvider
	fallback   Service // edge-tts fallback
}

// NewDoubaoTTSService 创建豆包 TTS 服务
func NewDoubaoTTSService(storageDir string, provider TTSSettingsProvider, fallback Service) *DoubaoTTSService {
	os.MkdirAll(storageDir, 0755)
	return &DoubaoTTSService{
		storageDir: storageDir,
		provider:   provider,
		fallback:   fallback,
	}
}

// Synthesize 文本转语音，返回文件名
func (s *DoubaoTTSService) Synthesize(ctx context.Context, text string) (string, error) {
	settings := s.provider.GetTTSSettings()

	// 如果配置为 edge 或豆包配置不完整，使用 fallback
	if settings.Provider != "doubao" || settings.AppID == "" || settings.AccessKey == "" || settings.SpeakerID == "" {
		if s.fallback != nil {
			return s.fallback.Synthesize(ctx, text)
		}
		return "", fmt.Errorf("doubao TTS not configured and no fallback available")
	}

	// 用 text + speakerID 生成缓存 key
	hash := md5.Sum([]byte(text + "|" + settings.SpeakerID))
	filename := "doubao_" + hex.EncodeToString(hash[:]) + ".mp3"
	fullPath := filepath.Join(s.storageDir, filename)

	// 缓存命中
	if _, err := os.Stat(fullPath); err == nil {
		return filename, nil
	}

	// 调用豆包 API
	audioData, err := s.callAPI(ctx, text, settings)
	if err != nil {
		logger.Log.Errorf("Doubao TTS API failed: %v, falling back to edge-tts", err)
		if s.fallback != nil {
			return s.fallback.Synthesize(ctx, text)
		}
		return "", fmt.Errorf("doubao TTS failed: %w", err)
	}

	if err := os.WriteFile(fullPath, audioData, 0644); err != nil {
		return "", fmt.Errorf("write audio file: %w", err)
	}

	return filename, nil
}

// callAPI 调用豆包 HTTP Chunked TTS 接口
func (s *DoubaoTTSService) callAPI(ctx context.Context, text string, settings TTSSettings) ([]byte, error) {
	reqBody := doubaoRequest{
		User: doubaoUser{UID: "linkflow"},
		ReqParams: doubaoReqParams{
			Text:    text,
			Speaker: settings.SpeakerID,
			AudioParams: doubaoAudioParams{
				Format:     "mp3",
				SampleRate: 24000,
			},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, doubaoTTSURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-App-Id", settings.AppID)
	req.Header.Set("X-Api-Access-Key", settings.AccessKey)
	req.Header.Set("X-Api-Resource-Id", settings.ResourceID)
	req.Header.Set("X-Api-Request-Id", generateRequestID())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	// 读取 chunked JSON 流，拼接 base64 音频数据
	var audioBuf bytes.Buffer
	scanner := bufio.NewScanner(resp.Body)
	// 增大 buffer 以处理较大的 base64 音频块
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var chunk doubaoChunk
		if err := json.Unmarshal(line, &chunk); err != nil {
			continue // 跳过非 JSON 行
		}

		// code=20000000 表示合成结束
		if chunk.Code == 20000000 {
			break
		}

		if chunk.Code != 0 {
			if chunk.Code >= 40000000 {
				return nil, fmt.Errorf("doubao API error: code=%d, msg=%s", chunk.Code, chunk.Message)
			}
			continue
		}

		// 解码 base64 音频数据
		if chunk.Data != "" {
			decoded, err := base64.StdEncoding.DecodeString(chunk.Data)
			if err != nil {
				logger.Log.Warnf("Doubao TTS: base64 decode error: %v", err)
				continue
			}
			audioBuf.Write(decoded)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if audioBuf.Len() == 0 {
		return nil, fmt.Errorf("no audio data received")
	}

	return audioBuf.Bytes(), nil
}

// generateRequestID 生成 UUID v4 格式的请求 ID
func generateRequestID() string {
	var b [16]byte
	rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
