package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ldchengyi/linkflow/internal/cache"
	"github.com/ldchengyi/linkflow/internal/config"
	"github.com/ldchengyi/linkflow/internal/database"
	"github.com/ldchengyi/linkflow/internal/handler"
	"github.com/ldchengyi/linkflow/internal/logger"
	"github.com/ldchengyi/linkflow/internal/middleware"
	mqttbroker "github.com/ldchengyi/linkflow/internal/mqtt"
	"github.com/ldchengyi/linkflow/internal/repository"
	"github.com/ldchengyi/linkflow/internal/scheduler"
	"github.com/ldchengyi/linkflow/internal/service"
	"github.com/ldchengyi/linkflow/internal/ws"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	// 初始化日志
	if err := logger.Init(logger.Config{
		Level:      cfg.Log.Level,
		Filename:   cfg.Log.Filename,
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
		Compress:   cfg.Log.Compress,
	}); err != nil {
		panic("Failed to init logger: " + err.Error())
	}
	defer logger.Sync()

	// 初始化数据库连接池
	ctx := context.Background()
	db, err := database.New(ctx, cfg.Database)
	if err != nil {
		logger.Log.Fatalf("Failed to connect database: %v", err)
	}
	defer db.Close()
	logger.Log.Info("Database connected")

	// 初始化 Redis
	rdb, err := cache.NewRedis(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		logger.Log.Fatalf("Failed to connect redis: %v", err)
	}
	defer rdb.Close()
	logger.Log.Info("Redis connected")

	// 初始化 WebSocket Hub
	hub := ws.NewHub()
	go hub.Run()
	logger.Log.Info("WebSocket hub started")

	// 初始化服务
	userRepo := repository.NewUserRepository(db.App())
	thingModelRepo := repository.NewThingModelRepository(db.App())
	jwtService := service.NewJWTService(cfg.JWT.Secret, cfg.JWT.ExpireHours, rdb)
	authService := service.NewAuthService(userRepo, jwtService)

	deviceRepo := repository.NewDeviceRepository(db.App())
	deviceDataRepoApp := repository.NewDeviceDataRepository(db.App())
	moduleRepo := repository.NewModuleRepository(db.App())
	auditLogRepo := repository.NewAuditLogRepository(db.Admin())
	alertRuleRepoApp := repository.NewAlertRuleRepository(db.App())
	alertLogRepoApp := repository.NewAlertLogRepository(db.App())
	scheduledTaskRepoApp := repository.NewScheduledTaskRepository(db.App())
	scheduledTaskRepoAdmin := repository.NewScheduledTaskRepository(db.Admin())
	stlRepoAdmin := repository.NewScheduledTaskLogRepository(db.Admin())
	stlRepoApp := repository.NewScheduledTaskLogRepository(db.App())

	// MQTT 专用 Repository（使用 Admin pool 绕过 RLS）
	mqttDeviceRepo := repository.NewDeviceRepository(db.Admin())
	mqttThingModelRepo := repository.NewThingModelRepository(db.Admin())
	deviceDataRepo := repository.NewDeviceDataRepository(db.Admin())
	mqttAlertRuleRepo := repository.NewAlertRuleRepository(db.Admin())
	mqttAlertLogRepo := repository.NewAlertLogRepository(db.Admin())
	mqttOTATaskRepo := repository.NewOTATaskRepository(db.Admin())
	mqttFirmwareRepo := repository.NewFirmwareRepository(db.Admin())

	// OTA Repository（App pool，受 RLS 约束）
	firmwareRepoApp := repository.NewFirmwareRepository(db.App())
	otaTaskRepoApp := repository.NewOTATaskRepository(db.App())

	// 平台设置 Repository（Admin pool，无 RLS）
	settingsRepo := repository.NewSettingsRepository(db.Admin())

	// 初始化 MQTT Broker
	baseURL := "http://localhost:" + cfg.Server.Port
	broker := mqttbroker.NewBroker(cfg.MQTT, mqttDeviceRepo, mqttThingModelRepo, deviceDataRepo, auditLogRepo, mqttAlertRuleRepo, mqttAlertLogRepo, mqttOTATaskRepo, mqttFirmwareRepo, settingsRepo, rdb, hub, baseURL)
	if err := broker.Start(); err != nil {
		logger.Log.Fatalf("Failed to start MQTT broker: %v", err)
	}
	logger.Log.Info("MQTT broker started")

	// 初始化 Scheduler
	sched := scheduler.New(scheduledTaskRepoAdmin, broker, stlRepoAdmin)
	sched.Start()

	// 初始化 Handler
	authHandler := handler.NewAuthHandler(authService)
	thingModelHandler := handler.NewThingModelHandler(thingModelRepo, db.App())
	deviceHandler := handler.NewDeviceHandler(deviceRepo, deviceDataRepoApp, db.App(), rdb)
	statsHandler := handler.NewStatsHandler(deviceRepo, thingModelRepo, alertLogRepoApp, db.App(), rdb)
	moduleHandler := handler.NewModuleHandler(moduleRepo)
	auditLogHandler := handler.NewAuditLogHandler(auditLogRepo)
	wsHandler := handler.NewWSHandler(hub, jwtService)
	alertRuleHandler := handler.NewAlertRuleHandler(alertRuleRepoApp, db.App(), broker)
	alertLogHandler := handler.NewAlertLogHandler(alertLogRepoApp, db.App())
	scheduledTaskHandler := handler.NewScheduledTaskHandler(scheduledTaskRepoApp, db.App())
	stlHandler := handler.NewScheduledTaskLogHandler(stlRepoApp, db.App())
	debugHandler := handler.NewDebugHandler(deviceRepo, thingModelRepo, deviceDataRepo, db.App(), rdb, broker)
	firmwareHandler := handler.NewFirmwareHandler(firmwareRepoApp, mqttDeviceRepo, db.App())
	otaTaskHandler := handler.NewOTATaskHandler(otaTaskRepoApp, firmwareRepoApp, deviceRepo, db.App(), broker, baseURL)
	settingsHandler := handler.NewSettingsHandler(settingsRepo, broker)

	// 设置路由
	router := gin.Default()

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, handler.Response{Code: 200, Msg: "success", Data: gin.H{"status": "ok"}})
	})

	// API 路由
	api := router.Group("/api")
	api.Use(middleware.AuditLog(auditLogRepo))
	{
		// 公开路由
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// WebSocket（token 通过 query param 认证，不走 Auth 中间件）
		api.GET("/ws", wsHandler.Connect)

		// 固件下载（设备通过 Basic Auth 认证，不走 JWT）
		api.GET("/firmwares/:id/download", firmwareHandler.Download)

		// 需要认证的路由
		protected := api.Group("")
		protected.Use(middleware.Auth(jwtService))
		{
			protected.POST("/auth/logout", authHandler.Logout)
			protected.PUT("/auth/password", authHandler.ChangePassword)
			protected.GET("/me", func(c *gin.Context) {
				c.JSON(http.StatusOK, handler.Response{Code: 200, Msg: "success", Data: gin.H{
					"user_id": middleware.GetUserID(c),
					"role":    middleware.GetUserRole(c),
				}})
			})

			// 物模型路由
			thingModels := protected.Group("/thing-models")
			{
				thingModels.POST("", thingModelHandler.Create)
				thingModels.GET("", thingModelHandler.List)
				thingModels.GET("/:id", thingModelHandler.Get)
				thingModels.PUT("/:id", thingModelHandler.Update)
				thingModels.DELETE("/:id", thingModelHandler.Delete)
			}

			// 设备路由
			devices := protected.Group("/devices")
			{
				devices.POST("", deviceHandler.Create)
				devices.GET("", deviceHandler.List)
				devices.GET("/:id", deviceHandler.Get)
				devices.GET("/:id/data/latest", deviceHandler.LatestData)
				devices.GET("/:id/data/history", deviceHandler.History)
				devices.GET("/:id/data/export", deviceHandler.ExportCSV)
				devices.PUT("/:id", deviceHandler.Update)
				devices.DELETE("/:id", deviceHandler.Delete)
				devices.POST("/:id/debug", debugHandler.Debug)
				devices.GET("/:id/connection-type", debugHandler.ConnectionType)
				devices.POST("/:id/simulate/online", debugHandler.SimulateOnline)
				devices.POST("/:id/simulate/offline", debugHandler.SimulateOffline)
				devices.POST("/:id/simulate/heartbeat", debugHandler.SimulateHeartbeat)
			}

			// 统计路由
			protected.GET("/stats/overview", statsHandler.Overview)

			// 模块路由
			modules := protected.Group("/modules")
			{
				modules.GET("", moduleHandler.List)
				modules.GET("/:id", moduleHandler.Get)
			}

			// 告警规则路由
			alertRules := protected.Group("/alert-rules")
			{
				alertRules.POST("", alertRuleHandler.Create)
				alertRules.GET("", alertRuleHandler.List)
				alertRules.GET("/:id", alertRuleHandler.Get)
				alertRules.PUT("/:id", alertRuleHandler.Update)
				alertRules.DELETE("/:id", alertRuleHandler.Delete)
			}

			// 告警日志路由（注意 unread-count 必须在 /:id 之前注册）
			protected.GET("/alert-logs/unread-count", alertLogHandler.UnreadCount)
			protected.GET("/alert-logs", alertLogHandler.List)
			protected.PUT("/alert-logs/:id/acknowledge", alertLogHandler.Acknowledge)

			// 定时任务路由
			scheduledTasks := protected.Group("/scheduled-tasks")
			{
				scheduledTasks.POST("", scheduledTaskHandler.Create)
				scheduledTasks.GET("", scheduledTaskHandler.List)
				scheduledTasks.GET("/:id", scheduledTaskHandler.Get)
				scheduledTasks.PUT("/:id", scheduledTaskHandler.Update)
				scheduledTasks.DELETE("/:id", scheduledTaskHandler.Delete)
			}

			// 定时任务执行日志路由
			protected.GET("/scheduled-task-logs", stlHandler.List)

			// 审计日志路由（admin 查所有，普通用户查自己的）
			protected.GET("/audit-logs", auditLogHandler.List)

			// 固件管理路由
			firmwares := protected.Group("/firmwares")
			{
				firmwares.POST("", firmwareHandler.Upload)
				firmwares.GET("", firmwareHandler.List)
				firmwares.DELETE("/:id", firmwareHandler.Delete)
			}

			// OTA 任务路由
			otaTasks := protected.Group("/ota-tasks")
			{
				otaTasks.POST("", otaTaskHandler.Create)
				otaTasks.GET("", otaTaskHandler.List)
				otaTasks.GET("/:id", otaTaskHandler.Get)
				otaTasks.PUT("/:id/cancel", otaTaskHandler.Cancel)
			}

			// 平台设置路由
			protected.GET("/settings", settingsHandler.Get)
			protected.PUT("/settings", settingsHandler.Update)
		}
	}

	// 启动服务器
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	go func() {
		logger.Log.Infof("Server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatalf("Server failed: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 先停 Scheduler 和 MQTT broker
	sched.Stop()
	if err := broker.Stop(ctx); err != nil {
		logger.Log.Errorf("MQTT broker shutdown error: %v", err)
	}

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Log.Info("Server exited")
}
