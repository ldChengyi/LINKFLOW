//go:build integration

package tests

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAPI_ModuleList(t *testing.T) {
	token := registerTestUser(t, "api_mod@linkflow.dev", "Test@12345")

	w := doRequest("GET", "/api/modules", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("list status = %d, body: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	var modules []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	json.Unmarshal(resp.Data, &modules)
	if len(modules) == 0 {
		t.Skip("no modules in database, skipping detail test")
	}

	// Get first module
	w = doRequest("GET", "/api/modules/"+modules[0].ID, nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("get status = %d", w.Code)
	}
}
