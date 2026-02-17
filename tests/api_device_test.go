//go:build integration

package tests

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAPI_DeviceCRUD(t *testing.T) {
	token := registerTestUser(t, "api_dev@linkflow.dev", "Test@12345")

	// Create
	w := doRequest("POST", "/api/devices", map[string]interface{}{
		"name": "测试设备",
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	var created struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		DeviceSecret string `json:"device_secret"`
	}
	json.Unmarshal(resp.Data, &created)
	if created.ID == "" || created.DeviceSecret == "" {
		t.Fatal("device ID or secret is empty")
	}
	devID := created.ID

	// Get
	w = doRequest("GET", "/api/devices/"+devID, nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("get status = %d", w.Code)
	}

	// List
	w = doRequest("GET", "/api/devices?page=1&page_size=10", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("list status = %d", w.Code)
	}
	resp = parseResp(t, w)
	pg := parsePage(t, resp.Data)
	if pg.Total < 1 {
		t.Error("expected at least 1 device")
	}

	// Update
	w = doRequest("PUT", "/api/devices/"+devID, map[string]interface{}{
		"name": "更新设备",
	}, token)
	if w.Code != http.StatusOK {
		t.Fatalf("update status = %d, body: %s", w.Code, w.Body.String())
	}

	// LatestData (no data yet, should still succeed)
	w = doRequest("GET", "/api/devices/"+devID+"/data/latest", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("latest data status = %d", w.Code)
	}

	// History
	w = doRequest("GET", "/api/devices/"+devID+"/data/history", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("history status = %d", w.Code)
	}

	// Delete
	w = doRequest("DELETE", "/api/devices/"+devID, nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("delete status = %d", w.Code)
	}

	// Verify deleted
	w = doRequest("GET", "/api/devices/"+devID, nil, token)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", w.Code)
	}
}
