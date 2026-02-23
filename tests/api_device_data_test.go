//go:build integration

package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// TestAPI_DeviceDataHistory 测试历史遥测数据接口的原始/聚合两种路径
func TestAPI_DeviceDataHistory(t *testing.T) {
	token := registerTestUser(t, "api_devdata@linkflow.dev", "Test@12345")

	// 创建测试设备
	w := doRequest("POST", "/api/devices", map[string]interface{}{
		"name": "数据聚合测试设备",
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create device: %d %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	var dev struct {
		ID string `json:"id"`
	}
	json.Unmarshal(resp.Data, &dev)
	devID := dev.ID

	// 获取当前用户 ID
	wme := doRequest("GET", "/api/me", nil, token)
	meResp := parseResp(t, wme)
	var meData struct {
		UserID string `json:"user_id"`
	}
	json.Unmarshal(meResp.Data, &meData)
	userID := meData.UserID

	// 插入测试遥测数据（每 10 分钟一条，覆盖最近 4 小时，共 24 条）
	now := time.Now().UTC()
	ctx := context.Background()
	for i := 0; i < 24; i++ {
		ts := now.Add(-time.Duration(i*10) * time.Minute)
		temp := 20 + i%5
		humi := 60 + i%10
		payload := fmt.Sprintf(`{"temperature": %d, "humidity": %d}`, temp, humi)
		_, err := testPool.Exec(ctx,
			`INSERT INTO device_data (time, device_id, user_id, topic, payload, qos, valid)
			 VALUES ($1, $2, $3, $4, $5::jsonb, 0, true)`,
			ts, devID, userID,
			fmt.Sprintf("devices/%s/telemetry/up", devID),
			payload,
		)
		if err != nil {
			t.Fatalf("insert telemetry row %d: %v", i, err)
		}
	}

	// histResp 解析历史接口统一响应
	type histResp struct {
		Aggregated bool            `json:"aggregated"`
		Interval   string          `json:"interval"`
		Data       json.RawMessage `json:"data"`
	}

	// ── 子测试1：1h 范围 → 原始数据 ────────────────────────────────────
	t.Run("raw_1h", func(t *testing.T) {
		start := now.Add(-1 * time.Hour).Format(time.RFC3339)
		end := now.Format(time.RFC3339)
		w := doRequest("GET",
			fmt.Sprintf("/api/devices/%s/data/history?start=%s&end=%s", devID, start, end),
			nil, token)
		if w.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
		outer := parseResp(t, w)
		var hr histResp
		if err := json.Unmarshal(outer.Data, &hr); err != nil {
			t.Fatalf("parse: %v (body=%s)", err, w.Body.String())
		}
		if hr.Aggregated {
			t.Error("want aggregated=false for 1h range")
		}
		if hr.Interval != "" {
			t.Errorf("want interval='', got '%s'", hr.Interval)
		}

		var rawPoints []struct {
			Time    string                 `json:"time"`
			Payload map[string]interface{} `json:"payload"`
			Valid   bool                   `json:"valid"`
		}
		if err := json.Unmarshal(hr.Data, &rawPoints); err != nil {
			t.Fatalf("parse raw points: %v", err)
		}
		// 过去 1h 内每 10 分钟一条，应有 ~6 条
		if len(rawPoints) == 0 {
			t.Error("expected at least 1 raw data point in 1h range")
		}
		t.Logf("raw_1h: %d points", len(rawPoints))
	})

	// ── 子测试2：6h 范围 → 5分钟聚合 ────────────────────────────────────
	t.Run("aggregated_6h", func(t *testing.T) {
		start := now.Add(-6 * time.Hour).Format(time.RFC3339)
		end := now.Format(time.RFC3339)
		w := doRequest("GET",
			fmt.Sprintf("/api/devices/%s/data/history?start=%s&end=%s", devID, start, end),
			nil, token)
		if w.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
		outer := parseResp(t, w)
		var hr histResp
		if err := json.Unmarshal(outer.Data, &hr); err != nil {
			t.Fatalf("parse: %v", err)
		}
		if !hr.Aggregated {
			t.Error("want aggregated=true for 6h range")
		}
		if hr.Interval != "5 minutes" {
			t.Errorf("want interval='5 minutes', got '%s'", hr.Interval)
		}

		var aggPoints []struct {
			Time       string             `json:"time"`
			Payload    map[string]float64 `json:"payload"`
			MaxPayload map[string]float64 `json:"max_payload"`
			MinPayload map[string]float64 `json:"min_payload"`
		}
		if err := json.Unmarshal(hr.Data, &aggPoints); err != nil {
			t.Fatalf("parse agg points: %v", err)
		}
		if len(aggPoints) == 0 {
			t.Fatal("expected at least 1 aggregated point in 6h range")
		}

		// 验证第一个点包含三组字段，且 max >= avg >= min
		pt := aggPoints[0]
		avg, avgOk := pt.Payload["temperature"]
		max, maxOk := pt.MaxPayload["temperature"]
		min, minOk := pt.MinPayload["temperature"]
		if !avgOk || !maxOk || !minOk {
			t.Fatalf("aggregated point missing temperature field: payload=%v max=%v min=%v",
				pt.Payload, pt.MaxPayload, pt.MinPayload)
		}
		if max < avg {
			t.Errorf("max(%.2f) < avg(%.2f)", max, avg)
		}
		if min > avg {
			t.Errorf("min(%.2f) > avg(%.2f)", min, avg)
		}
		t.Logf("aggregated_6h: %d points, sample temp avg=%.2f max=%.2f min=%.2f",
			len(aggPoints), avg, max, min)
	})

	// ── 子测试3：7d 范围 → 1小时聚合 ────────────────────────────────────
	t.Run("aggregated_7d", func(t *testing.T) {
		start := now.Add(-7 * 24 * time.Hour).Format(time.RFC3339)
		end := now.Format(time.RFC3339)
		w := doRequest("GET",
			fmt.Sprintf("/api/devices/%s/data/history?start=%s&end=%s", devID, start, end),
			nil, token)
		if w.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
		outer := parseResp(t, w)
		var hr histResp
		if err := json.Unmarshal(outer.Data, &hr); err != nil {
			t.Fatalf("parse: %v", err)
		}
		if !hr.Aggregated {
			t.Error("want aggregated=true for 7d range")
		}
		if hr.Interval != "1 hour" {
			t.Errorf("want interval='1 hour', got '%s'", hr.Interval)
		}
		t.Logf("aggregated_7d: interval=%s", hr.Interval)
	})

	// ── 子测试4：30d 范围 → 6小时聚合 ────────────────────────────────────
	t.Run("aggregated_30d", func(t *testing.T) {
		start := now.Add(-30 * 24 * time.Hour).Format(time.RFC3339)
		end := now.Format(time.RFC3339)
		w := doRequest("GET",
			fmt.Sprintf("/api/devices/%s/data/history?start=%s&end=%s", devID, start, end),
			nil, token)
		if w.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
		outer := parseResp(t, w)
		var hr histResp
		if err := json.Unmarshal(outer.Data, &hr); err != nil {
			t.Fatalf("parse: %v", err)
		}
		if !hr.Aggregated {
			t.Error("want aggregated=true for 30d range")
		}
		if hr.Interval != "6 hours" {
			t.Errorf("want interval='6 hours', got '%s'", hr.Interval)
		}
		t.Logf("aggregated_30d: interval=%s", hr.Interval)
	})

	// ── 子测试5：无数据时返回空数组而非报错 ──────────────────────────────
	t.Run("empty_result", func(t *testing.T) {
		// 查询未来 1h（肯定没数据）
		start := now.Add(1 * time.Hour).Format(time.RFC3339)
		end := now.Add(2 * time.Hour).Format(time.RFC3339)
		w := doRequest("GET",
			fmt.Sprintf("/api/devices/%s/data/history?start=%s&end=%s", devID, start, end),
			nil, token)
		if w.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
	})

	// 清理
	testPool.Exec(ctx, "DELETE FROM device_data WHERE device_id = $1", devID)
	doRequest("DELETE", "/api/devices/"+devID, nil, token)
}
