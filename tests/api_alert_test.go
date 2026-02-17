//go:build integration

package tests

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAPI_AlertRuleCRUD(t *testing.T) {
	token := registerTestUser(t, "api_alert@linkflow.dev", "Test@12345")

	// Create a device first (alert rules need a device_id)
	w := doRequest("POST", "/api/devices", map[string]interface{}{
		"name": "告警测试设备",
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create device: %d %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	var dev struct{ ID string `json:"id"` }
	json.Unmarshal(resp.Data, &dev)

	// Create alert rule
	w = doRequest("POST", "/api/alert-rules", map[string]interface{}{
		"name":        "温度过高",
		"device_id":   dev.ID,
		"property_id": "temp",
		"operator":    ">",
		"threshold":   50,
		"severity":    "warning",
		"enabled":     true,
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create rule: %d %s", w.Code, w.Body.String())
	}
	resp = parseResp(t, w)
	var rule struct{ ID string `json:"id"` }
	json.Unmarshal(resp.Data, &rule)
	ruleID := rule.ID

	// Get
	w = doRequest("GET", "/api/alert-rules/"+ruleID, nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("get rule: %d", w.Code)
	}

	// List
	w = doRequest("GET", "/api/alert-rules?page=1&page_size=10", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("list rules: %d", w.Code)
	}
	resp = parseResp(t, w)
	pg := parsePage(t, resp.Data)
	if pg.Total < 1 {
		t.Error("expected at least 1 alert rule")
	}

	// List filtered by device_id
	w = doRequest("GET", "/api/alert-rules?device_id="+dev.ID, nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("list rules by device: %d", w.Code)
	}

	// Update
	w = doRequest("PUT", "/api/alert-rules/"+ruleID, map[string]interface{}{
		"name":        "温度过高-更新",
		"device_id":   dev.ID,
		"property_id": "temp",
		"operator":    ">=",
		"threshold":   60,
		"severity":    "critical",
		"enabled":     true,
	}, token)
	if w.Code != http.StatusOK {
		t.Fatalf("update rule: %d %s", w.Code, w.Body.String())
	}

	// Delete
	w = doRequest("DELETE", "/api/alert-rules/"+ruleID, nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("delete rule: %d", w.Code)
	}
}

func TestAPI_AlertLogList(t *testing.T) {
	token := registerTestUser(t, "api_alertlog@linkflow.dev", "Test@12345")

	w := doRequest("GET", "/api/alert-logs?page=1&page_size=10", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("list alert logs: %d %s", w.Code, w.Body.String())
	}
}
