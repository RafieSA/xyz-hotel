package handler

import (
	"encoding/json"
	"strconv"
	"strings"

	"xyz-hotel/backend/internal/model"
	"xyz-hotel/backend/internal/repo"
	"xyz-hotel/backend/internal/ws"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// ChatHandler handles POST /api/chat and GET /api/chat.
type ChatHandler struct {
	MessageRepo *repo.MessageRepo
	BookingRepo *repo.BookingRepo
}

func NewChatHandler(mr *repo.MessageRepo, br *repo.BookingRepo) *ChatHandler {
	return &ChatHandler{MessageRepo: mr, BookingRepo: br}
}

type ChatRequest struct {
	Message   string `json:"message"`
	BookingID *int64 `json:"booking_id"`
}

func isAdminRole(role string) bool {
	return role == model.RoleOwner || role == model.RoleManager || role == model.RoleReceptionist
}

// PostChat handles POST /api/chat {message, booking_id?} Auth required.
// Customer can only chat with own booking_id else 403, admin any.
func (h *ChatHandler) PostChat(c *fiber.Ctx) error {
	userID, ok := getAuthUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Please sign in to send messages"})
	}
	role, _ := c.Locals("role").(string)
	var req ChatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "We could not read your request. Check the format and try again"})
	}
	msg := strings.TrimSpace(req.Message)
	if len(msg) < 1 || len(msg) > 500 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Message must be 1 to 500 characters"})
	}
	if req.BookingID != nil {
		bid := *req.BookingID
		if bid <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Booking ID is invalid"})
		}
		booking, err := h.BookingRepo.GetByID(bid)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Booking not found"})
		}
		if !isAdminRole(role) && booking.UserID != userID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "You can only send messages for your own bookings"})
		}
	}
	isAdmin := isAdminRole(role)
	m := &model.Message{
		UserID:    userID,
		BookingID: req.BookingID,
		Message:   msg,
		IsAdmin:   isAdmin,
	}
	// If admin posts for another user's booking, we keep admin's user_id but note booking_id.
	// For admin replying to a customer without booking_id, target is admin's own user_id loop; but admin inbox filtering handles it.
	// To support admin replying to customer user, frontend may pass implicit target via booking. For generic admin->user, we keep as is.
	created, err := h.MessageRepo.Create(m)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not send your message. Please try again"})
	}
	// Broadcast via WS hub
	go func() {
		data, _ := json.Marshal(map[string]interface{}{"type": "chat_message", "message": created})
		ws.GetHub().Broadcast(data)
	}()
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Message sent", "data": created})
}

// GetChat handles GET /api/chat?booking_id? Auth required, customer own only, admin all.
func (h *ChatHandler) GetChat(c *fiber.Ctx) error {
	userID, ok := getAuthUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Please sign in to view messages"})
	}
	role, _ := c.Locals("role").(string)
	admin := isAdminRole(role)
	bidStr := c.Query("booking_id")
	if bidStr != "" {
		bid, err := strconv.ParseInt(bidStr, 10, 64)
		if err != nil || bid <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Booking ID is invalid"})
		}
		if !admin {
			booking, err := h.BookingRepo.GetByID(bid)
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Booking not found"})
			}
			if booking.UserID != userID {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "You can only view messages for your own bookings"})
			}
		}
		list, err := h.MessageRepo.ListByBooking(bid)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load messages. Please try again"})
		}
		return c.JSON(fiber.Map{"data": list})
	}
	if admin {
		list, err := h.MessageRepo.ListForAdmin(nil)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load messages. Please try again"})
		}
		return c.JSON(fiber.Map{"data": list})
	}
	list, err := h.MessageRepo.ListByUser(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load messages. Please try again"})
	}
	return c.JSON(fiber.Map{"data": list})
}

// GetConversations handles GET /api/chat/conversations (admin)
func (h *ChatHandler) GetConversations(c *fiber.Ctx) error {
	role, _ := c.Locals("role").(string)
	if !isAdminRole(role) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "You do not have permission for this"})
	}
	list, err := h.MessageRepo.GetConversations()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load conversations. Please try again"})
	}
	return c.JSON(fiber.Map{"data": list})
}

// GetAdminMessages handles GET /api/admin/chat/messages?user_id? for inbox
func (h *ChatHandler) GetAdminMessages(c *fiber.Ctx) error {
	uidStr := c.Query("user_id")
	if uidStr != "" {
		uid, err := strconv.ParseInt(uidStr, 10, 64)
		if err != nil || uid <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "User ID is invalid"})
		}
		list, err := h.MessageRepo.ListForAdmin(&uid)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load messages. Please try again"})
		}
		return c.JSON(fiber.Map{"data": list})
	}
	list, err := h.MessageRepo.ListForAdmin(nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load messages. Please try again"})
	}
	return c.JSON(fiber.Map{"data": list})
}

// WsChat upgrade handler for /ws/chat
func WsChat(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// WsChatHandler is websocket handler after upgrade, reuses global hub.
func WsChatHandler(c *websocket.Conn) {
	hub := ws.GetHub()
	hub.Register(c)
	defer func() {
		hub.Unregister(c)
		_ = c.Close()
	}()
	for {
		mt, msg, err := c.ReadMessage()
		if err != nil {
			break
		}
		if mt == websocket.TextMessage && len(msg) > 0 {
			// Broadcast chat echo if client sends JSON; otherwise pong
			var payload map[string]interface{}
			if json.Unmarshal(msg, &payload) == nil {
				// If valid JSON, broadcast as is with type chat_message if missing
				if _, ok := payload["type"]; !ok {
					payload["type"] = "chat_message"
					if data, err := json.Marshal(payload); err == nil {
						hub.Broadcast(data)
						continue
					}
				} else {
					hub.Broadcast(msg)
					continue
				}
			}
			_ = c.WriteMessage(websocket.TextMessage, []byte(`{"type":"pong"}`))
		}
	}
}
