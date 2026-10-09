package handler

import (
	"database/sql"
	"strconv"
	"strings"

	"xyz-hotel/backend/internal/model"
	"xyz-hotel/backend/internal/repo"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// RoomHandler handles admin CRUD for room_types and room_units.
type RoomHandler struct {
	RoomRepo  *repo.RoomRepo
	Validator *validator.Validate
}

func NewRoomHandler(r *repo.RoomRepo) *RoomHandler {
	return &RoomHandler{RoomRepo: r, Validator: validator.New()}
}

type createRoomTypeReq struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"max=1000"`
	Capacity    int    `json:"capacity" validate:"required,gte=1,lte=10"`
	Price       int64  `json:"price" validate:"required,gte=0"`
	TotalUnits  int    `json:"total_units" validate:"gte=0"`
}

type updateRoomTypeReq struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"max=1000"`
	Capacity    int    `json:"capacity" validate:"required,gte=1,lte=10"`
	Price       int64  `json:"price" validate:"required,gte=0"`
	TotalUnits  int    `json:"total_units" validate:"gte=0"`
}

// ListRoomTypesAdmin handles GET /api/admin/room-types
func (h *RoomHandler) ListRoomTypesAdmin(c *fiber.Ctx) error {
	list, err := h.RoomRepo.ListRoomTypesAdmin()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load room types. Please try again"})
	}
	return c.JSON(fiber.Map{"data": list})
}

// ListRoomTypesPublic handles GET /api/room-types?q=deluxe (public, no auth) filtering ILIKE name/description where deleted_at IS NULL
func (h *RoomHandler) ListRoomTypesPublic(c *fiber.Ctx) error {
	q := strings.TrimSpace(c.Query("q"))
	if q != "" {
		list, err := h.RoomRepo.SearchRoomTypesWithRating(q)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load room types. Please try again"})
		}
		return c.JSON(fiber.Map{"data": list})
	}
	list, err := h.RoomRepo.ListRoomTypesWithRating()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load room types. Please try again"})
	}
	return c.JSON(fiber.Map{"data": list})
}
// CreateRoomType handles POST /api/admin/room-types (owner/manager)
func (h *RoomHandler) CreateRoomType(c *fiber.Ctx) error {
	var req createRoomTypeReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "We could not read your request. Check the format and try again"})
	}
	if err := h.Validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Please check your room type details", "details": err.Error()})
	}
	rt := &model.RoomType{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Capacity:    req.Capacity,
		Price:       req.Price,
		TotalUnits:  req.TotalUnits,
	}
	created, err := h.RoomRepo.CreateRoomType(rt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": "A room type with that name already exists"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not create the room type. Please try again"})
	}
	// audit
	userID, _ := getAuthUserID(c)
	if userID != 0 {
		_, _ = h.RoomRepo.DB.Exec(`INSERT INTO audit_logs (user_id, action, entity, entity_id, payload) VALUES ($1,$2,$3,$4,$5::jsonb)`, userID, "room_type.create", "room_types", created.ID, `{"name":"`+created.Name+`"}`)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Room type created", "data": created})
}

// UpdateRoomType handles PUT /api/admin/room-types/:id
func (h *RoomHandler) UpdateRoomType(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Room type ID is invalid"})
	}
	var req updateRoomTypeReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "We could not read your request. Check the format and try again"})
	}
	if err := h.Validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Please check your room type details", "details": err.Error()})
	}
	rt := &model.RoomType{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Capacity:    req.Capacity,
		Price:       req.Price,
		TotalUnits:  req.TotalUnits,
	}
	updated, err := h.RoomRepo.UpdateRoomType(id, rt)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Room type not found"})
		}
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": "A room type with that name already exists"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not update the room type. Please try again"})
	}
	userID, _ := getAuthUserID(c)
	if userID != 0 {
		_, _ = h.RoomRepo.DB.Exec(`INSERT INTO audit_logs (user_id, action, entity, entity_id, payload) VALUES ($1,$2,$3,$4,$5::jsonb)`, userID, "room_type.update", "room_types", id, `{"name":"`+updated.Name+`"}`)
	}
	return c.JSON(fiber.Map{"message": "Room type updated", "data": updated})
}

// DeleteRoomType handles DELETE /api/admin/room-types/:id (soft delete)
func (h *RoomHandler) DeleteRoomType(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Room type ID is invalid"})
	}
	if err := h.RoomRepo.SoftDeleteRoomType(id); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Room type not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not delete the room type. Please try again"})
	}
	userID, _ := getAuthUserID(c)
	if userID != 0 {
		_, _ = h.RoomRepo.DB.Exec(`INSERT INTO audit_logs (user_id, action, entity, entity_id, payload) VALUES ($1,$2,$3,$4,$5::jsonb)`, userID, "room_type.delete", "room_types", id, `{}`)
	}
	return c.JSON(fiber.Map{"message": "Room type deleted"})
}

// --- Room Units CRUD ---

type createRoomUnitReq struct {
	RoomTypeID int64  `json:"room_type_id" validate:"required,gt=0"`
	Code       string `json:"code" validate:"required,min=2,max=50"`
	Status     string `json:"status" validate:"omitempty,oneof=available occupied dirty maintenance"`
}

// ListRoomUnits handles GET /api/admin/room-units
func (h *RoomHandler) ListRoomUnits(c *fiber.Ctx) error {
	list, err := h.RoomRepo.ListAllUnits(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load room units. Please try again"})
	}
	return c.JSON(fiber.Map{"data": list})
}

// CreateRoomUnit handles POST /api/admin/room-units
func (h *RoomHandler) CreateRoomUnit(c *fiber.Ctx) error {
	var req createRoomUnitReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "We could not read your request. Check the format and try again"})
	}
	if err := h.Validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Please check room unit details", "details": err.Error()})
	}
	status := req.Status
	if status == "" {
		status = model.RoomStatusAvailable
	}
	code := strings.TrimSpace(req.Code)
	// verify room_type exists
	if _, err := h.RoomRepo.GetRoomTypeByID(req.RoomTypeID); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Room type not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not verify room type. Please try again"})
	}
	created, err := h.RoomRepo.CreateRoomUnit(req.RoomTypeID, code, status)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": "A room unit with that code already exists"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not create the room unit. Please try again"})
	}
	userID, _ := getAuthUserID(c)
	if userID != 0 {
		_, _ = h.RoomRepo.DB.Exec(`INSERT INTO audit_logs (user_id, action, entity, entity_id, payload) VALUES ($1,$2,$3,$4,$5::jsonb)`, userID, "room_unit.create", "room_units", created.ID, `{"code":"`+created.Code+`"}`)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Room unit created", "data": created})
}

// DeleteRoomUnit handles DELETE /api/admin/room-units/:id
func (h *RoomHandler) DeleteRoomUnit(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Room unit ID is invalid"})
	}
	if err := h.RoomRepo.SoftDeleteRoomUnit(id); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Room unit not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not delete the room unit. Please try again"})
	}
	return c.JSON(fiber.Map{"message": "Room unit deleted"})
}
