package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DifyCommand Dify 工作流返回的指令结构
type DifyCommand struct {
	Action     string         `json:"action"`       // set_property | invoke_service | query_status | unknown
	PropertyID string         `json:"property_id"`
	Value      any            `json:"value"`
	ServiceID  string         `json:"service_id"`
	Params     map[string]any `json:"params"`
	Message    string         `json:"message"`
}

// difyRequest Dify workflow/run 请求体
type difyRequest struct {
	Inputs       map[string]string `json:"inputs"`
	ResponseMode string            `json:"response_mode"`
	User         string            `json:"user"`
}

// difyResponse Dify API 响应（blocking 模式）
type difyResponse struct {
	Data struct {
		Outputs map[string]any `json:"outputs"`
		Status  string         `json:"status"`
		Error   string         `json:"error"`
	} `json:"data"`
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

// CallWorkflow 调用 Dify 工作流，返回解析后的 DifyCommand
func CallWorkflow(apiURL, apiKey, deviceContext, userInput, deviceID string) (*DifyCommand, error) {
	reqBody := difyRequest{
		Inputs: map[string]string{
			"device_context": deviceContext,
			"user_input":     userInput,
		},
		ResponseMode: "blocking",
		User:         fmt.Sprintf("linkflow_%s", deviceID),
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/workflows/run", apiURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("dify returned %d: %s", resp.StatusCode, string(body))
	}

	var difyResp difyResponse
	if err := json.NewDecoder(resp.Body).Decode(&difyResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if difyResp.Data.Status == "failed" {
		return nil, fmt.Errorf("dify workflow failed: %s", difyResp.Data.Error)
	}

	// 从 outputs.result 提取 JSON 字符串
	resultRaw, ok := difyResp.Data.Outputs["result"]
	if !ok {
		return nil, fmt.Errorf("dify response missing outputs.result")
	}

	var resultStr string
	switch v := resultRaw.(type) {
	case string:
		resultStr = v
	default:
		b, _ := json.Marshal(v)
		resultStr = string(b)
	}

	var cmd DifyCommand
	if err := json.Unmarshal([]byte(resultStr), &cmd); err != nil {
		return nil, fmt.Errorf("parse DifyCommand from result: %w", err)
	}

	return &cmd, nil
}
