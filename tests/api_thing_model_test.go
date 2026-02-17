//go:build integration

package tests

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAPI_ThingModelCRUD(t *testing.T) {
	token := registerTestUser(t, "api_tm@linkflow.dev", "Test@12345")

	// Create
	createBody := map[string]interface{}{
		"name":        "测试物模型",
		"description": "API测试用",
		"properties": []map[string]interface{}{
			{"id": "temp", "name": "温度", "dataType": "float", "unit": "°C", "accessMode": "r", "min": -40, "max": 80},
			{"id": "switch", "name": "开关", "dataType": "bool", "accessMode": "rw"},
		},
		"events":   []map[string]interface{}{},
		"services": []map[string]interface{}{},
		"modules":  []map[string]interface{}{},
	}
	w := doRequest("POST", "/api/thing-models", createBody, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	var created struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	json.Unmarshal(resp.Data, &created)
	if created.ID == "" {
		t.Fatal("created thing model ID is empty")
	}
	tmID := created.ID

	// Get
	w = doRequest("GET", "/api/thing-models/"+tmID, nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("get status = %d", w.Code)
	}

	// List
	w = doRequest("GET", "/api/thing-models?page=1&page_size=10", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("list status = %d", w.Code)
	}
	resp = parseResp(t, w)
	pg := parsePage(t, resp.Data)
	if pg.Total < 1 {
		t.Error("expected at least 1 thing model")
	}

	// Update
	updateBody := map[string]interface{}{
		"name":        "更新物模型",
		"description": "已更新",
		"properties": []map[string]interface{}{
			{"id": "temp", "name": "温度", "dataType": "float", "unit": "°C", "accessMode": "r"},
		},
		"events":   []map[string]interface{}{},
		"services": []map[string]interface{}{},
		"modules":  []map[string]interface{}{},
	}
	w = doRequest("PUT", "/api/thing-models/"+tmID, updateBody, token)
	if w.Code != http.StatusOK {
		t.Fatalf("update status = %d, body: %s", w.Code, w.Body.String())
	}

	// Delete
	w = doRequest("DELETE", "/api/thing-models/"+tmID, nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("delete status = %d", w.Code)
	}

	// Verify deleted
	w = doRequest("GET", "/api/thing-models/"+tmID, nil, token)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", w.Code)
	}
}
