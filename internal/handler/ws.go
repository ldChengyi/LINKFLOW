package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ldchengyi/linkflow/internal/logger"
	"github.com/ldchengyi/linkflow/internal/service"
	"github.com/ldchengyi/linkflow/internal/ws"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WSHandler WebSocket 连接处理器
type WSHandler struct {
	hub        *ws.Hub
	jwtService *service.JWTService
}

func NewWSHandler(hub *ws.Hub, jwtService *service.JWTService) *WSHandler {
	return &WSHandler{hub: hub, jwtService: jwtService}
}

// Connect 升级 HTTP 为 WebSocket（通过 query param token 认证）
func (h *WSHandler) Connect(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, Response{Code: 401, Msg: "missing token"})
		return
	}

	claims, err := h.jwtService.Validate(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, Response{Code: 401, Msg: err.Error()})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Log.Errorf("WS upgrade failed: %v", err)
		return
	}

	ws.ServeClient(h.hub, claims.UserID, conn)
}
