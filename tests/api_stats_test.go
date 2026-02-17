//go:build integration

package tests

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAPI_StatsOverview(t *testing.T) {
	token := registerTestUser(t, "api_stats@linkflow.dev", "Test@12345")

	w := doRequest("GET", "/api/stats/overview", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	var stats struct {
		TotalDevices     int   `json:"total_devices"`
		OnlineDevices    int64 `json:"online_devices"`
		TotalThingModels int   `json:"total_thing_models"`
	}
	json.Unmarshal(resp.Data, &stats)
	if stats.TotalDevices < 0 || stats.TotalThingModels < 0 {
		t.Errorf("unexpected stats: %+v", stats)
	}
}
