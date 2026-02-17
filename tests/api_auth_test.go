//go:build integration

package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestAPI_Register(t *testing.T) {
	email := "api_register@linkflow.dev"
	testPool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email)

	w := doRequest("POST", "/api/auth/register", map[string]string{
		"email": email, "password": "Test@12345",
	}, "")
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	var data struct {
		Token string `json:"token"`
		User  struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"user"`
	}
	json.Unmarshal(resp.Data, &data)
	if data.Token == "" || data.User.Email != email {
		t.Errorf("unexpected data: %+v", data)
	}
}

func TestAPI_RegisterDuplicate(t *testing.T) {
	email := "api_dup@linkflow.dev"
	testPool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email)

	doRequest("POST", "/api/auth/register", map[string]string{
		"email": email, "password": "Test@12345",
	}, "")

	w := doRequest("POST", "/api/auth/register", map[string]string{
		"email": email, "password": "Test@12345",
	}, "")
	if w.Code == http.StatusCreated {
		t.Fatal("expected duplicate register to fail")
	}
}

func TestAPI_Login(t *testing.T) {
	email := "api_login@linkflow.dev"
	testPool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email)

	doRequest("POST", "/api/auth/register", map[string]string{
		"email": email, "password": "Test@12345",
	}, "")

	w := doRequest("POST", "/api/auth/login", map[string]string{
		"email": email, "password": "Test@12345",
	}, "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	var data struct{ Token string `json:"token"` }
	json.Unmarshal(resp.Data, &data)
	if data.Token == "" {
		t.Error("token is empty")
	}
}

func TestAPI_LoginWrongPassword(t *testing.T) {
	email := "api_wrongpw@linkflow.dev"
	testPool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email)

	doRequest("POST", "/api/auth/register", map[string]string{
		"email": email, "password": "Test@12345",
	}, "")

	w := doRequest("POST", "/api/auth/login", map[string]string{
		"email": email, "password": "WrongPassword",
	}, "")
	if w.Code == http.StatusOK {
		t.Fatal("expected login with wrong password to fail")
	}
}

func TestAPI_Logout(t *testing.T) {
	token := registerTestUser(t, "api_logout@linkflow.dev", "Test@12345")

	w := doRequest("POST", "/api/auth/logout", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("logout status = %d, want 200", w.Code)
	}

	// token should be revoked
	w = doRequest("GET", "/api/me", nil, token)
	if w.Code == http.StatusOK {
		t.Fatal("expected revoked token to be rejected")
	}
}

func TestAPI_Me(t *testing.T) {
	token := registerTestUser(t, "api_me@linkflow.dev", "Test@12345")

	w := doRequest("GET", "/api/me", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	resp := parseResp(t, w)
	var data struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}
	json.Unmarshal(resp.Data, &data)
	if data.UserID == "" {
		t.Error("user_id is empty")
	}
}

func TestAPI_Unauthorized(t *testing.T) {
	w := doRequest("GET", "/api/me", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}
