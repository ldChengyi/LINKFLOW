//go:build integration

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ldchengyi/linkflow/internal/handler"
	"github.com/ldchengyi/linkflow/internal/middleware"
	"github.com/ldchengyi/linkflow/internal/repository"
	"github.com/ldchengyi/linkflow/internal/service"
)

var (
	testRouter     *gin.Engine
	testJWTService *service.JWTService
)

func setupRouter() {
	gin.SetMode(gin.TestMode)

	userRepo := repository.NewUserRepository(testPool)
	thingModelRepo := repository.NewThingModelRepository(testPool)
	deviceRepo := repository.NewDeviceRepository(testPool)
	deviceDataRepo := repository.NewDeviceDataRepository(testPool)
	moduleRepo := repository.NewModuleRepository(testPool)
	auditLogRepo := repository.NewAuditLogRepository(testPool)
	alertRuleRepo := repository.NewAlertRuleRepository(testPool)
	alertLogRepo := repository.NewAlertLogRepository(testPool)
	scheduledTaskRepo := repository.NewScheduledTaskRepository(testPool)

	jwtCfg := testCfg.Integration.JWT
	testJWTService = service.NewJWTService(jwtCfg.Secret, jwtCfg.ExpireHours, testRdb)
	authService := service.NewAuthService(userRepo, testJWTService)

	authHandler := handler.NewAuthHandler(authService)
	thingModelHandler := handler.NewThingModelHandler(thingModelRepo, testPool)
	deviceHandler := handler.NewDeviceHandler(deviceRepo, deviceDataRepo, testPool, testRdb)
	statsHandler := handler.NewStatsHandler(deviceRepo, thingModelRepo, testPool, testRdb)
	moduleHandler := handler.NewModuleHandler(moduleRepo)
	alertRuleHandler := handler.NewAlertRuleHandler(alertRuleRepo, testPool, nil)
	alertLogHandler := handler.NewAlertLogHandler(alertLogRepo, testPool)
	scheduledTaskHandler := handler.NewScheduledTaskHandler(scheduledTaskRepo, testPool)

	r := gin.New()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, handler.Response{Code: 200, Msg: "success", Data: gin.H{"status": "ok"}})
	})

	api := r.Group("/api")
	api.Use(middleware.AuditLog(auditLogRepo))
	{
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)

		p := api.Group("")
		p.Use(middleware.Auth(testJWTService))
		{
			p.POST("/auth/logout", authHandler.Logout)
			p.GET("/me", func(c *gin.Context) {
				c.JSON(200, handler.Response{Code: 200, Msg: "success", Data: gin.H{
					"user_id": middleware.GetUserID(c),
					"role":    middleware.GetUserRole(c),
				}})
			})

			tm := p.Group("/thing-models")
			{
				tm.POST("", thingModelHandler.Create)
				tm.GET("", thingModelHandler.List)
				tm.GET("/:id", thingModelHandler.Get)
				tm.PUT("/:id", thingModelHandler.Update)
				tm.DELETE("/:id", thingModelHandler.Delete)
			}

			d := p.Group("/devices")
			{
				d.POST("", deviceHandler.Create)
				d.GET("", deviceHandler.List)
				d.GET("/:id", deviceHandler.Get)
				d.GET("/:id/data/latest", deviceHandler.LatestData)
				d.GET("/:id/data/history", deviceHandler.History)
				d.PUT("/:id", deviceHandler.Update)
				d.DELETE("/:id", deviceHandler.Delete)
			}

			p.GET("/stats/overview", statsHandler.Overview)

			mod := p.Group("/modules")
			{
				mod.GET("", moduleHandler.List)
				mod.GET("/:id", moduleHandler.Get)
			}

			ar := p.Group("/alert-rules")
			{
				ar.POST("", alertRuleHandler.Create)
				ar.GET("", alertRuleHandler.List)
				ar.GET("/:id", alertRuleHandler.Get)
				ar.PUT("/:id", alertRuleHandler.Update)
				ar.DELETE("/:id", alertRuleHandler.Delete)
			}

			p.GET("/alert-logs", alertLogHandler.List)

			st := p.Group("/scheduled-tasks")
			{
				st.POST("", scheduledTaskHandler.Create)
				st.GET("", scheduledTaskHandler.List)
				st.GET("/:id", scheduledTaskHandler.Get)
				st.PUT("/:id", scheduledTaskHandler.Update)
				st.DELETE("/:id", scheduledTaskHandler.Delete)
			}
		}
	}

	testRouter = r
}

// apiResp 统一响应解析
type apiResp struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

// pageResp 分页响应解析
type pageResp struct {
	List     json.RawMessage `json:"list"`
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

func doRequest(method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}
	req, _ := http.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

func parseResp(t *testing.T, w *httptest.ResponseRecorder) apiResp {
	t.Helper()
	var resp apiResp
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v, body: %s", err, w.Body.String())
	}
	return resp
}

func parsePage(t *testing.T, data json.RawMessage) pageResp {
	t.Helper()
	var p pageResp
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatalf("parse page: %v", err)
	}
	return p
}

// registerTestUser 注册测试用户并返回 token
func registerTestUser(t *testing.T, email, password string) string {
	t.Helper()
	// 先清理
	testPool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email)

	w := doRequest("POST", "/api/auth/register", map[string]string{
		"email": email, "password": password,
	}, "")
	if w.Code != http.StatusCreated {
		t.Fatalf("register failed: %d %s", w.Code, w.Body.String())
	}
	resp := parseResp(t, w)
	var data struct{ Token string `json:"token"` }
	json.Unmarshal(resp.Data, &data)
	return data.Token
}
