//go:build integration

package tests

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAPI_ScheduledTaskCRUD(t *testing.T) {
	token := registerTestUser(t, "api_sched@linkflow.dev", "Test@12345")

	// Create a device first
	w := doRequest("POST", "/api/devices", map[string]interface{}{
		"name": "定时任务测试设备",
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create device: %d %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	var dev struct{ ID string `json:"id"` }
	json.Unmarshal(resp.Data, &dev)

	// Create scheduled task
	w = doRequest("POST", "/api/scheduled-tasks", map[string]interface{}{
		"device_id":   dev.ID,
		"name":        "每分钟开灯",
		"cron_expr":   "* * * * *",
		"action_type": "property_set",
		"property_id": "switch",
		"value":       true,
		"enabled":     true,
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create task: %d %s", w.Code, w.Body.String())
	}
	resp = parseResp(t, w)
	var task struct{ ID string `json:"id"` }
	json.Unmarshal(resp.Data, &task)
	taskID := task.ID

	// Get
	w = doRequest("GET", "/api/scheduled-tasks/"+taskID, nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("get task: %d", w.Code)
	}

	// List
	w = doRequest("GET", "/api/scheduled-tasks?page=1&page_size=10", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("list tasks: %d", w.Code)
	}
	resp = parseResp(t, w)
	pg := parsePage(t, resp.Data)
	if pg.Total < 1 {
		t.Error("expected at least 1 scheduled task")
	}

	// List filtered by device_id
	w = doRequest("GET", "/api/scheduled-tasks?device_id="+dev.ID, nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("list by device: %d", w.Code)
	}

	// Update
	w = doRequest("PUT", "/api/scheduled-tasks/"+taskID, map[string]interface{}{
		"device_id":   dev.ID,
		"name":        "每小时开灯",
		"cron_expr":   "0 * * * *",
		"action_type": "property_set",
		"property_id": "switch",
		"value":       false,
		"enabled":     false,
	}, token)
	if w.Code != http.StatusOK {
		t.Fatalf("update task: %d %s", w.Code, w.Body.String())
	}

	// Delete
	w = doRequest("DELETE", "/api/scheduled-tasks/"+taskID, nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("delete task: %d", w.Code)
	}
}
