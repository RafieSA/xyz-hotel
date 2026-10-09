package handler

import (
	"log/slog"

	"xyz-hotel/backend/internal/ws"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// WsAdmin handles GET /ws/admin websocket upgrades (Auth+RBAC). Keeps connection alive and broadcasts.
func WsAdmin(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// WsAdminHandler is the websocket handler after upgrade.
func WsAdminHandler(c *websocket.Conn) {
	hub := ws.GetHub()
	hub.Register(c)
	slog.Info("ws admin connected", "remote", c.RemoteAddr().String())
	defer func() {
		hub.Unregister(c)
		_ = c.Close()
		slog.Info("ws admin disconnected")
	}()
	for {
		mt, msg, err := c.ReadMessage()
		if err != nil {
			break
		}
		// echo or ping handling, keep alive
		if mt == websocket.TextMessage && len(msg) > 0 {
			// optional: ignore client messages, just keep connection
			_ = c.WriteMessage(websocket.TextMessage, []byte(`{"type":"pong"}`))
		}
	}
}
