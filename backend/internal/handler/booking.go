package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"xyz-hotel/backend/internal/model"
	"xyz-hotel/backend/internal/repo"
	"xyz-hotel/backend/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// BookingHandler handles availability and bookings.
type BookingHandler struct {
	Availability *service.AvailabilityService
	BookingRepo  *repo.BookingRepo
	BookingOps   *service.BookingOpsService
	Validator    *validator.Validate
}

func NewBookingHandler(svc *service.AvailabilityService, br *repo.BookingRepo) *BookingHandler {
	return &BookingHandler{
		Availability: svc,
		BookingRepo:  br,
		Validator:    validator.New(),
	}
}

// NewBookingHandlerWithOps creates handler with ops injected.
func NewBookingHandlerWithOps(svc *service.AvailabilityService, br *repo.BookingRepo, ops *service.BookingOpsService) *BookingHandler {
	return &BookingHandler{
		Availability: svc,
		BookingRepo:  br,
		BookingOps:   ops,
		Validator:    validator.New(),
	}
}

// ListBookingsStub placeholder for Fase 1 (kept for compatibility, not used).
func ListBookingsStub(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"data":    []interface{}{},
		"message": "stub - not implemented yet",
	})
}

// CreateBookingStub placeholder for Fase 1 (kept for compatibility, not used).
func CreateBookingStub(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"message": "stub - not implemented yet",
	})
}

// GetAvailability handles GET /api/availability?room_type_id&check_in&check_out (public).
func (h *BookingHandler) GetAvailability(c *fiber.Ctx) error {
	roomTypeIDStr := c.Query("room_type_id")
	checkIn := c.Query("check_in")
	checkOut := c.Query("check_out")

	if roomTypeIDStr == "" || checkIn == "" || checkOut == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Room type, check in and check out dates are required. Use YYYY-MM-DD for dates"})
	}
	roomTypeID, err := strconv.ParseInt(roomTypeIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Room type is invalid. Please choose a valid room type"})
	}

	result, err := h.Availability.CheckAvailability(c.Context(), roomTypeID, checkIn, checkOut)
	if err != nil {
		msg := err.Error()
		if contains(msg, "Room type not found") || contains(msg, "room type not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": msg})
		}
		if contains(msg, "Check out must be after") || contains(msg, "check_out must be after") || contains(msg, "Check in date is invalid") || contains(msg, "Check out date is invalid") || contains(msg, "invalid check") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
		}
		slog.Error("availability check failed", "err", err, "room_type_id", roomTypeID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not check availability. Please try again"})
	}
	return c.JSON(fiber.Map{
		"data": result,
	})
}
// CreateBookingRequest validation struct.
type CreateBookingRequest struct {
	RoomTypeID  int64  `json:"room_type_id" validate:"required"`
	CheckIn     string `json:"check_in" validate:"required"`
	CheckOut    string `json:"check_out" validate:"required"`
	Guests      int    `json:"guests" validate:"required,gte=1"`
	VoucherCode string `json:"voucher_code" validate:"omitempty"`
}

// CreateBooking handles POST /api/bookings (auth required).
func (h *BookingHandler) CreateBooking(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	userID, ok := userIDVal.(int64)
	if !ok {
		// fallback for int / float64 from JWT
		switch v := userIDVal.(type) {
		case int:
			userID = int64(v)
			ok = true
		case int64:
			ok = true
		case float64:
			userID = int64(v)
			ok = true
		}
	}
	if !ok || userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Please sign in to book"})
	}

	var req CreateBookingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "We could not read your request. Check the format and try again"})
	}
	if err := h.Validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Please check your booking details and try again", "details": err.Error()})
	}
	booking, err := h.Availability.CreateBooking(c.Context(), userID, req.RoomTypeID, req.CheckIn, req.CheckOut, req.Guests, req.VoucherCode)
	if err != nil {
		msg := err.Error()
		if contains(msg, "Room type not found") || contains(msg, "room type not found") || contains(msg, "Voucher code not found") || contains(msg, "voucher not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": msg})
		}
		if contains(msg, "No rooms available") || contains(msg, "no available units") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": msg})
		}
		if contains(msg, "expired") || contains(msg, "usage limit") || contains(msg, "quota exceeded") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
		}
		if contains(msg, "Check out must be after") || contains(msg, "check_out must be after") || contains(msg, "Check in date is invalid") || contains(msg, "Check out date is invalid") || contains(msg, "invalid check") || contains(msg, "Guests must be") || contains(msg, "guests must be") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
		}
		// capacity exceeded etc.
		if contains(msg, "Too many guests") || contains(msg, "guests exceeds") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
		}
		if contains(msg, "needs at least") || contains(msg, "voucher requires minimum") || contains(msg, "minimum") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
		}
		if contains(msg, "voucher") || contains(msg, "Voucher") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
		}
		slog.Error("create booking failed", "err", err, "user_id", userID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not create your booking. Please try again"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": booking})
}
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}

// ListBookings handles GET /api/bookings (auth required, own or admin).
// Customers see only own bookings (IDOR prevention). Owner/manager/receptionist see all.
func (h *BookingHandler) ListBookings(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	roleVal, _ := c.Locals("role").(string)
	var userID int64
	switch v := userIDVal.(type) {
	case int64:
		userID = v
	case int:
		userID = int64(v)
	case float64:
		userID = int64(v)
	default:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Please sign in to view bookings"})
	}

	isAdmin := roleVal == "owner" || roleVal == "manager" || roleVal == "receptionist"
	if isAdmin {
		// BFLA: admin can see all, but still logged
		list, err := h.BookingRepo.ListAll()
		if err != nil {
			slog.Error("list all bookings failed", "err", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load bookings. Please try again"})
		}
		return c.JSON(fiber.Map{"data": list})
	}
	// Customer: only own bookings (BOLA/IDOR prevention: ignore any query param user_id)
	list, err := h.BookingRepo.ListByUser(userID)
	if err != nil {
		slog.Error("list user bookings failed", "err", err, "user_id", userID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load your bookings. Please try again"})
	}
	return c.JSON(fiber.Map{"data": list})
}

// --- Proof upload ---

const maxProofSize = 5 * 1024 * 1024 // 5MB

var allowedExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".pdf":  true,
}

// getAuthUserID extracts user_id from context with int/float fallbacks.
func getAuthUserID(c *fiber.Ctx) (int64, bool) {
	val := c.Locals("user_id")
	switch v := val.(type) {
	case int64:
		if v != 0 {
			return v, true
		}
	case int:
		if v != 0 {
			return int64(v), true
		}
	case float64:
		if v != 0 {
			return int64(v), true
		}
	}
	return 0, false
}

// ValidateProofFile validates extension, size and magic bytes.
// Exported for unit testing.
func ValidateProofFile(filename string, size int64, header []byte) error {
	if size > maxProofSize {
		return fmt.Errorf("File is too large. Maximum size is 5MB")
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedExts[ext] {
		return fmt.Errorf("Unsupported file type. Use JPG, PNG or PDF")
	}
	if len(header) == 0 {
		return fmt.Errorf("File is empty. Please choose a file with content")
	}
	// MIME via magic
	isJPEG := len(header) >= 3 && header[0] == 0xFF && header[1] == 0xD8 && header[2] == 0xFF
	isPNG := len(header) >= 8 && header[0] == 0x89 && header[1] == 0x50 && header[2] == 0x4E && header[3] == 0x47 && header[4] == 0x0D && header[5] == 0x0A && header[6] == 0x1A && header[7] == 0x0A
	isPDF := len(header) >= 4 && header[0] == 0x25 && header[1] == 0x50 && header[2] == 0x44 && header[3] == 0x46 // %PDF
	// ext based magic check: pdf must be PDF, jpg must be JPEG, png must be PNG
	switch ext {
	case ".jpg", ".jpeg":
		if !isJPEG {
			return fmt.Errorf("File does not look like a valid JPG. Please upload a real JPG image")
		}
	case ".png":
		if !isPNG {
			return fmt.Errorf("File does not look like a valid PNG. Please upload a real PNG image")
		}
	case ".pdf":
		if !isPDF {
			return fmt.Errorf("File does not look like a valid PDF. Please upload a real PDF file")
		}
	}
	// generic fallback if none matched (should not happen)
	if !(isJPEG || isPNG || isPDF) {
		return fmt.Errorf("Only JPG, PNG and PDF files are allowed")
	}
	return nil
}
// UploadProof handles POST /api/bookings/:id/proof multipart/form-data proof file.
// Auth required, IDOR check booking.user_id == auth.id, size <=5MB, ext + magic validation, save to storage/uploads/bookings/{id}_{uuid}.ext
func (h *BookingHandler) UploadProof(c *fiber.Ctx) error {
	userID, ok := getAuthUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Please sign in to upload proof"})
	}
	idStr := c.Params("id")
	bookingID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || bookingID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Booking ID is invalid"})
	}
	if h.BookingRepo == nil || h.BookingRepo.DB == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"message": "Service is temporarily unavailable. Please try again later"})
	}
	booking, err := h.BookingRepo.GetByID(bookingID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Booking not found"})
	}
	// IDOR check
	if booking.UserID != userID {
		slog.Warn("idor blocked: proof upload", "booking_id", bookingID, "owner", booking.UserID, "requester", userID)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "You can only upload proof for your own bookings"})
	}
	if booking.Status != model.BookingPendingPayment {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "This booking cannot accept proof right now. It must be waiting for payment"})
	}

	fileHeader, err := c.FormFile("proof")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Please attach your payment proof in the proof field"})
	}
	// Enforce size before reading
	if fileHeader.Size > maxProofSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "File is too large. Maximum size is 5MB"})
	}
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowedExts[ext] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Unsupported file type. Use JPG, PNG or PDF"})
	}

	// Read header for magic validation (first 512 bytes)
	f, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "We could not read your file. Please try again"})
	}
	defer f.Close()
	header := make([]byte, 512)
	n, _ := io.ReadFull(f, header)
	if n > 0 {
		header = header[:n]
	} else {
		header = header[:0]
	}
	if err := ValidateProofFile(fileHeader.Filename, fileHeader.Size, header); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	// Prepare save path bookings/{id}_{uuid}.ext
	relPath := fmt.Sprintf("bookings/%d_%s%s", bookingID, uuid.NewString(), ext)
	// storage base: backend/storage/uploads relative to working dir; also fallback to absolute
	baseDir := "storage/uploads"
	// If running from different cwd, ensure bookings subdir exists
	fullPath := filepath.Join(baseDir, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		slog.Error("mkdir uploads failed", "err", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not save your file. Please try again"})
	}
	// Save file: reopen to read from start
	_ = f.Close()
	f2, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not read your file. Please try again"})
	}
	defer f2.Close()
	out, err := os.Create(fullPath)
	if err != nil {
		slog.Error("create proof file failed", "err", err, "path", fullPath)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not save your file. Please try again"})
	}
	defer out.Close()
	if _, err := io.Copy(out, f2); err != nil {
		slog.Error("copy proof file failed", "err", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not save your file. Please try again"})
	}

	// Update booking proof_url + status pending_payment→waiting_verification
	updated, err := h.BookingRepo.UpdateProofURL(bookingID, relPath, model.BookingWaitingVerification)
	if err != nil {
		slog.Error("update proof_url failed", "err", err, "booking_id", bookingID)
		_ = os.Remove(fullPath)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not update your booking. Please try again"})
	}

	// Audit log
	payload, _ := json.Marshal(map[string]interface{}{
		"proof_url": relPath,
		"prev_status": model.BookingPendingPayment,
		"new_status": model.BookingWaitingVerification,
	})
	_, _ = h.BookingRepo.DB.Exec(`INSERT INTO audit_logs (user_id, action, entity, entity_id, payload) VALUES ($1,$2,$3,$4,$5::jsonb)`,
		userID, "booking.proof_upload", "bookings", bookingID, string(payload))
	slog.Info("proof uploaded", "booking_id", bookingID, "user_id", userID, "proof_url", relPath)

	return c.JSON(fiber.Map{"message": "Payment proof uploaded. We will verify it shortly", "proof_url": relPath, "data": updated})
}
// VerifyRequest body for PATCH /api/admin/bookings/:id/verify
type VerifyRequest struct {
	Action       string  `json:"action" validate:"required,oneof=verified rejected"`
	RejectReason *string `json:"reject_reason"`
}

// VerifyBooking handles PATCH /api/admin/bookings/:id/verify
// RBAC owner/manager only (enforced via middleware), checks status == waiting_verification
func (h *BookingHandler) VerifyBooking(c *fiber.Ctx) error {
	userID, ok := getAuthUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Please sign in to verify bookings"})
	}
	role, _ := c.Locals("role").(string)
	if role != model.RoleOwner && role != model.RoleManager {
		slog.Warn("bfla blocked: verify", "role", role, "user_id", userID)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "You do not have permission to verify bookings"})
	}
	idStr := c.Params("id")
	bookingID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || bookingID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Booking ID is invalid"})
	}
	if h.BookingRepo == nil || h.BookingRepo.DB == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"message": "Service is temporarily unavailable. Please try again later"})
	}
	var req VerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "We could not read your request. Check the format and try again"})
	}
	if err := h.Validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Please choose verified or rejected", "details": err.Error()})
	}
	if req.Action == "rejected" {
		if req.RejectReason == nil || strings.TrimSpace(*req.RejectReason) == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Please provide a reason when rejecting a booking"})
		}
	}
	booking, err := h.BookingRepo.GetByID(bookingID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Booking not found"})
	}
	if booking.Status != model.BookingWaitingVerification {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "This booking is not ready for verification. It must be waiting for verification"})
	}
	var newStatus string
	var rejectReason *string
	if req.Action == "verified" {
		newStatus = model.BookingVerified
	} else {
		newStatus = model.BookingRejected
		rr := strings.TrimSpace(*req.RejectReason)
		rejectReason = &rr
	}
	updated, err := h.BookingRepo.UpdateStatus(bookingID, newStatus, rejectReason)
	if err != nil {
		slog.Error("verify update failed", "err", err, "booking_id", bookingID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not update the booking. Please try again"})
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"action":        req.Action,
		"new_status":    newStatus,
		"reject_reason": rejectReason,
		"prev_status":   model.BookingWaitingVerification,
	})
	_, _ = h.BookingRepo.DB.Exec(`INSERT INTO audit_logs (user_id, action, entity, entity_id, payload) VALUES ($1,$2,$3,$4,$5::jsonb)`,
		userID, "booking.verify", "bookings", bookingID, string(payload))
	slog.Info("booking verified", "booking_id", bookingID, "action", req.Action, "by", userID, "role", role)
	if req.Action == "verified" {
		return c.JSON(fiber.Map{"message": "Booking verified successfully", "data": updated})
	}
	return c.JSON(fiber.Map{"message": "Booking rejected", "data": updated})
}

// CheckIn handles PATCH /api/admin/bookings/:id/checkin (owner/manager/receptionist)
func (h *BookingHandler) CheckIn(c *fiber.Ctx) error {
	userID, ok := getAuthUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Please sign in to check in guests"})
	}
	idStr := c.Params("id")
	bookingID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || bookingID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Booking ID is invalid"})
	}
	if h.BookingOps == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"message": "Service is temporarily unavailable. Please try again later"})
	}
	updated, err := h.BookingOps.CheckIn(c.Context(), bookingID, userID)
	if err != nil {
		msg := err.Error()
		switch {
		case contains(msg, "Booking not found") || msg == "booking not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Booking not found"})
		case contains(msg, "Only verified bookings") || contains(msg, "must be verified"):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
		case contains(msg, "No rooms available") || contains(msg, "no available units"):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": msg})
		case contains(msg, "already has a room") || contains(msg, "already has room"):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": msg})
		default:
			slog.Error("checkin failed", "err", err, "booking_id", bookingID)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not check in the guest. Please try again"})
		}
	}
	return c.JSON(fiber.Map{"message": "Guest checked in successfully", "data": updated})
}

// CheckOut handles PATCH /api/admin/bookings/:id/checkout (owner/manager/receptionist)
func (h *BookingHandler) CheckOut(c *fiber.Ctx) error {
	userID, ok := getAuthUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Please sign in to check out guests"})
	}
	idStr := c.Params("id")
	bookingID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || bookingID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Booking ID is invalid"})
	}
	if h.BookingOps == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"message": "Service is temporarily unavailable. Please try again later"})
	}
	updated, err := h.BookingOps.CheckOut(c.Context(), bookingID, userID)
	if err != nil {
		msg := err.Error()
		switch {
		case contains(msg, "Booking not found") || msg == "booking not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Booking not found"})
		case contains(msg, "Only checked in") || contains(msg, "must be checked_in"):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
		case contains(msg, "no room") || contains(msg, "has no room"):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
		default:
			slog.Error("checkout failed", "err", err, "booking_id", bookingID)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not check out the guest. Please try again"})
		}
	}
	return c.JSON(fiber.Map{"message": "Guest checked out successfully", "data": updated})
}

// UpdateRoomUnitStatusRequest for PATCH /api/admin/room-units/:id/status
type UpdateRoomUnitStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=available occupied dirty maintenance"`
}

// UpdateRoomUnitStatus handles PATCH /api/admin/room-units/:id/status (owner/manager/receptionist)
func (h *BookingHandler) UpdateRoomUnitStatus(c *fiber.Ctx) error {
	userID, ok := getAuthUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Please sign in to update room status"})
	}
	idStr := c.Params("id")
	unitID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || unitID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Room ID is invalid"})
	}
	var req UpdateRoomUnitStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "We could not read your request. Check the format and try again"})
	}
	if err := h.Validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Please choose available, occupied, dirty or maintenance", "details": err.Error()})
	}
	if h.BookingOps == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"message": "Service is temporarily unavailable. Please try again later"})
	}
	updated, err := h.BookingOps.UpdateRoomUnitStatus(c.Context(), unitID, req.Status, userID)
	if err != nil {
		msg := err.Error()
		switch {
		case contains(msg, "Room not found") || msg == "room unit not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": msg})
		case contains(msg, "Room is already") || contains(msg, "already"):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
		case contains(msg, "Cannot change room") || contains(msg, "invalid transition"):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
		case contains(msg, "Status is invalid") || contains(msg, "invalid status"):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
		default:
			slog.Error("update unit status failed", "err", err, "unit_id", unitID)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not update the room. Please try again"})
		}
	}
	return c.JSON(fiber.Map{"message": "Room status updated", "data": updated})
}
