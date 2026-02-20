//go:build integration

package tests

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAPI_FirmwareCRUD(t *testing.T) {
	token := registerTestUser(t, "api_fw@linkflow.dev", "Test@12345")

	// Upload firmware
	w := doMultipart("/api/firmwares", map[string]string{
		"name": "测试固件", "version": "1.0.0", "description": "单元测试固件",
	}, "test.bin", []byte("fake-firmware-binary-content"), token)
	if w.Code != http.StatusCreated {
		t.Fatalf("upload firmware: %d %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	var fw struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Version  string `json:"version"`
		Checksum string `json:"checksum"`
		FileSize int    `json:"file_size"`
	}
	json.Unmarshal(resp.Data, &fw)
	if fw.ID == "" {
		t.Fatal("firmware id is empty")
	}
	if fw.Checksum == "" {
		t.Error("expected checksum")
	}
	if fw.FileSize == 0 {
		t.Error("expected file_size > 0")
	}

	// List
	w = doRequest("GET", "/api/firmwares?page=1&page_size=10", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("list firmwares: %d", w.Code)
	}
	resp = parseResp(t, w)
	pg := parsePage(t, resp.Data)
	if pg.Total < 1 {
		t.Error("expected at least 1 firmware")
	}

	// Delete
	w = doRequest("DELETE", "/api/firmwares/"+fw.ID, nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("delete firmware: %d %s", w.Code, w.Body.String())
	}
}

func TestAPI_OTATaskCRUD(t *testing.T) {
	token := registerTestUser(t, "api_ota@linkflow.dev", "Test@12345")

	// Create device
	w := doRequest("POST", "/api/devices", map[string]interface{}{
		"name": "OTA测试设备",
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create device: %d %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	var dev struct{ ID string `json:"id"` }
	json.Unmarshal(resp.Data, &dev)

	// Upload firmware
	w = doMultipart("/api/firmwares", map[string]string{
		"name": "OTA固件", "version": "2.0.0",
	}, "ota.bin", []byte("ota-firmware-content"), token)
	if w.Code != http.StatusCreated {
		t.Fatalf("upload firmware: %d %s", w.Code, w.Body.String())
	}
	resp = parseResp(t, w)
	var fw struct{ ID string `json:"id"` }
	json.Unmarshal(resp.Data, &fw)

	// Create OTA task
	w = doRequest("POST", "/api/ota-tasks", map[string]interface{}{
		"device_id": dev.ID, "firmware_id": fw.ID,
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create ota task: %d %s", w.Code, w.Body.String())
	}
	resp = parseResp(t, w)
	var task struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	json.Unmarshal(resp.Data, &task)
	if task.ID == "" {
		t.Fatal("ota task id is empty")
	}

	// Get
	w = doRequest("GET", "/api/ota-tasks/"+task.ID, nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("get ota task: %d", w.Code)
	}

	// List
	w = doRequest("GET", "/api/ota-tasks?page=1&page_size=10", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("list ota tasks: %d", w.Code)
	}
	resp = parseResp(t, w)
	pg := parsePage(t, resp.Data)
	if pg.Total < 1 {
		t.Error("expected at least 1 ota task")
	}

	// List filtered by device_id
	w = doRequest("GET", "/api/ota-tasks?device_id="+dev.ID, nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("list by device: %d", w.Code)
	}

	// Cancel
	w = doRequest("PUT", "/api/ota-tasks/"+task.ID+"/cancel", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("cancel ota task: %d %s", w.Code, w.Body.String())
	}
}
