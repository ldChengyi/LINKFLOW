//go:build integration

package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/ldchengyi/linkflow/internal/model"
)

// voiceTestDevice 保存测试设备的基本信息
type voiceTestDevice struct {
	DeviceID     string
	DeviceSecret string
}

// setupVoiceTestData 创建测试用物模型 + 设备，返回设备信息
func setupVoiceTestData(t *testing.T, token string) voiceTestDevice {
	t.Helper()

	// 创建带有语音模块的物模型
	tmBody := map[string]interface{}{
		"name":        "语音测试物模型",
		"description": "集成测试专用",
		"properties": []map[string]interface{}{
			{"id": "switch", "name": "开关", "dataType": "bool", "accessMode": "rw"},
			{"id": "brightness", "name": "亮度", "dataType": "int", "accessMode": "rw", "min": 0, "max": 100, "step": 10, "unit": "%"},
			{"id": "mode", "name": "模式", "dataType": "enum", "accessMode": "rw",
				"enumValues": []map[string]interface{}{
					{"value": 0, "label": "正常"},
					{"value": 1, "label": "节能"},
					{"value": 2, "label": "强光"},
				}},
			{"id": "temperature", "name": "温度", "dataType": "float", "accessMode": "r", "unit": "℃"},
		},
		"services": []map[string]interface{}{
			{"id": "reboot", "name": "重启"},
		},
		"modules": []map[string]interface{}{
			{
				"id": "voice",
				"config": map[string]interface{}{
					"exposed_properties": []string{"switch", "brightness", "mode", "temperature"},
					"exposed_services":   []string{"reboot"},
				},
			},
		},
	}
	w := doRequest("POST", "/api/thing-models", tmBody, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create thing model failed: %d %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	var tm struct{ ID string `json:"id"` }
	json.Unmarshal(resp.Data, &tm)

	// 创建设备，名称含"测试灯"（用于语音指令中的设备名匹配）
	devBody := map[string]interface{}{
		"name":     "测试灯",
		"model_id": tm.ID,
	}
	w = doRequest("POST", "/api/devices", devBody, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create device failed: %d %s", w.Code, w.Body.String())
	}
	resp = parseResp(t, w)
	var dev struct {
		ID           string `json:"id"`
		DeviceSecret string `json:"device_secret"`
	}
	json.Unmarshal(resp.Data, &dev)

	// 预写入初始设备数据，供 query_status 和相对调节使用
	// 使用明确的 user_id 查询避免子查询失败
	ctx := context.Background()
	var userID string
	if err := testPool.QueryRow(ctx,
		"SELECT id FROM users WHERE email = 'voice_test@linkflow.dev'",
	).Scan(&userID); err != nil || userID == "" {
		t.Fatalf("get user id failed: %v", err)
	}
	initialPayload, _ := json.Marshal(map[string]interface{}{
		"switch":      false,
		"brightness":  50,
		"mode":        0,
		"temperature": 23.5,
	})
	if _, err := testPool.Exec(ctx, `
		INSERT INTO device_data (time, device_id, user_id, topic, payload, qos, valid)
		VALUES (NOW(), $1, $2, $3, $4, 1, true)
	`, dev.ID, userID, "devices/"+dev.ID+"/telemetry/up", initialPayload); err != nil {
		t.Fatalf("insert initial device_data failed: %v", err)
	}

	return voiceTestDevice{DeviceID: dev.ID, DeviceSecret: dev.DeviceSecret}
}

// mqttConnect 以设备身份连接到 MQTT Broker，返回 client 和 voice/down 结果 channel
func mqttConnect(t *testing.T, deviceID, deviceSecret string) (mqtt.Client, <-chan model.VoiceResult) {
	t.Helper()

	resultCh := make(chan model.VoiceResult, 8)
	downTopic := fmt.Sprintf("devices/%s/voice/down", deviceID)

	opts := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883").
		SetClientID(deviceID).
		SetUsername(deviceID).
		SetPassword(deviceSecret).
		SetConnectTimeout(10 * time.Second).
		SetAutoReconnect(false).
		SetDefaultPublishHandler(func(_ mqtt.Client, msg mqtt.Message) {
			var result model.VoiceResult
			if err := json.Unmarshal(msg.Payload(), &result); err == nil {
				resultCh <- result
			}
		})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); !token.WaitTimeout(10*time.Second) || token.Error() != nil {
		t.Fatalf("MQTT connect failed: %v", token.Error())
	}

	// 订阅 voice/down
	if token := client.Subscribe(downTopic, 1, nil); !token.WaitTimeout(5*time.Second) || token.Error() != nil {
		t.Fatalf("MQTT subscribe failed: %v", token.Error())
	}

	return client, resultCh
}

// sendVoice 发布语音指令，等待 voice/down 回传
func sendVoice(t *testing.T, client mqtt.Client, resultCh <-chan model.VoiceResult, deviceID, text string) model.VoiceResult {
	t.Helper()
	upTopic := fmt.Sprintf("devices/%s/voice/up", deviceID)
	payload, _ := json.Marshal(map[string]string{"text": text})
	token := client.Publish(upTopic, 1, false, payload)
	token.WaitTimeout(3 * time.Second)

	select {
	case result := <-resultCh:
		return result
	case <-time.After(5 * time.Second):
		t.Fatalf("timeout waiting for voice/down response (text=%q)", text)
		return model.VoiceResult{}
	}
}

func TestVoiceModule(t *testing.T) {
	// 注册测试用户
	token := registerTestUser(t, "voice_test@linkflow.dev", "Voice@12345")

	// 创建测试数据（物模型 + 设备 + 初始遥测数据）
	dev := setupVoiceTestData(t, token)
	t.Logf("Test device: id=%s", dev.DeviceID)

	// 连接 MQTT（以设备身份）
	client, resultCh := mqttConnect(t, dev.DeviceID, dev.DeviceSecret)
	defer client.Disconnect(500)

	// 等待 broker 完成 OnConnect 注册（写 devices sync.Map）
	time.Sleep(400 * time.Millisecond)

	tests := []struct {
		name    string
		text    string
		wantOK  bool
		wantIn  string // 期望 action 中包含的子串（空表示不验证）
		note    string // 说明
	}{
		// ── 开关控制（带设备名，触发 matchSingleBool）──
		{
			name: "bool_on_with_device",
			text: "打开测试灯",
			wantOK: true, wantIn: "switch",
			note: "设备名 + 开关词 → 模糊 bool 匹配",
		},
		{
			name: "bool_off_with_device",
			text: "关闭测试灯",
			wantOK: true, wantIn: "switch",
			note: "设备名 + 关闭词",
		},

		// ── 数值设定（带设备名 + 属性名 + 数值）──
		{
			name: "set_brightness",
			text: "测试灯亮度调到80",
			wantOK: true, wantIn: "brightness",
			note: "设备名 + 属性名 + 调到X",
		},

		// ── 相对调节（step=10，初始/当前值 50/80）──
		{
			name: "increase_brightness",
			text: "测试灯亮度调高",
			wantOK: true, wantIn: "brightness",
			note: "调高 → brightness + step(10)",
		},
		{
			name: "decrease_brightness",
			text: "测试灯亮度调低",
			wantOK: true, wantIn: "brightness",
			note: "调低 → brightness - step(10)",
		},

		// ── 枚举匹配（需要包含触发词 + 属性名 + 枚举 label）──
		// "节能模式" 单独使用会因意图分类器没有枚举关键词而失败
		// 正确用法：包含 property_set 触发词（"设为"）+ 属性名（"模式"）+ enum label（"节能"）
		{
			name: "enum_by_reverse_lookup",
			text: "设为节能模式",
			wantOK: true, wantIn: "mode",
			note: "设为(触发词) + 节能(enum label) + 模式(属性名) → 反向查找",
		},

		// ── 服务调用 ──
		{
			name: "service_reboot",
			text: "重启测试灯",
			wantOK: true, wantIn: "reboot",
			note: "服务名直接匹配",
		},

		// ── 状态查询（只读属性 temperature，依赖初始写入的数据）──
		{
			name: "query_temperature",
			text: "测试灯温度是多少",
			wantOK: true, wantIn: "temperature",
			note: "query_status 意图 + 只读属性",
		},
		{
			name: "query_brightness",
			text: "测试灯亮度当前是多少",
			wantOK: true, wantIn: "brightness",
			note: "query_status 意图 + 可读属性",
		},

		// ── 预期失败：意图无法识别 ──
		{
			name: "fail_no_keyword",
			text: "节能模式", // 无 property_set 触发词
			wantOK: false,
			note: "枚举 label 单独出现，意图分类失败",
		},
		{
			name: "fail_unrelated",
			text: "今天天气怎么样",
			wantOK: false,
			note: "完全无关指令",
		},

		// ── 预期失败：无设备名且反向查找无匹配 ──
		{
			name: "fail_no_device_name",
			text: "打开", // 无设备名、无属性名，反向查找找不到
			wantOK: false,
			note: "matchSingleBool 仅在已匹配设备名后触发",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := sendVoice(t, client, resultCh, dev.DeviceID, tc.text)
			t.Logf("  text=%q → success=%v | msg=%q | action=%q  [%s]",
				tc.text, result.Success, result.Message, result.Action, tc.note)

			if result.Success != tc.wantOK {
				t.Errorf("expected success=%v, got=%v (message=%q)",
					tc.wantOK, result.Success, result.Message)
			}
			if tc.wantOK && tc.wantIn != "" && !contains(result.Action, tc.wantIn) {
				t.Errorf("expected action to contain %q, got %q", tc.wantIn, result.Action)
			}
		})
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && func() bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}())
}
