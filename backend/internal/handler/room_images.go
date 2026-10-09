package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"xyz-hotel/backend/internal/repo"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const maxImageSize = 5 * 1024 * 1024 // 5MB

var allowedImageExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true,
}

// RoomImageHandler handles gallery upload/list/delete.
type RoomImageHandler struct {
	Repo *repo.RoomImageRepo
	RoomRepo *repo.RoomRepo
}

func NewRoomImageHandler(r *repo.RoomImageRepo, roomRepo *repo.RoomRepo) *RoomImageHandler {
	return &RoomImageHandler{Repo: r, RoomRepo: roomRepo}
}

// ListImages handles GET /api/room-types/:id/images public
func (h *RoomImageHandler) ListImages(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		idStr = c.Params("room_type_id")
	}
	roomTypeID, err := parseID(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid room type id"})
	}
	// verify room type exists (soft delete aware)
	if h.RoomRepo != nil {
		if _, err := h.RoomRepo.GetRoomTypeByID(roomTypeID); err != nil {
			// try admin fallback
			if _, err2 := h.RoomRepo.GetRoomTypeAdmin(roomTypeID); err2 != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Room type not found"})
			}
		}
	}
	images, err := h.Repo.List(roomTypeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load images"})
	}
	return c.JSON(fiber.Map{"data": images})
}

// UploadImage handles POST /api/admin/room-types/:id/images multipart
func (h *RoomImageHandler) UploadImage(c *fiber.Ctx) error {
	idStr := c.Params("id")
	roomTypeID, err := parseID(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid room type id"})
	}
	if h.RoomRepo != nil {
		if _, err := h.RoomRepo.GetRoomTypeAdmin(roomTypeID); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Room type not found"})
		}
	}
	// check count <5
	cnt, err := h.Repo.Count(roomTypeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not check image count"})
	}
	if cnt >= 5 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Max 5 photos per room type"})
	}
	file, err := c.FormFile("image")
	if err != nil {
		// try generic file field
		file, err = c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Image file is required. Use field 'image'"})
		}
	}
	if file.Size > maxImageSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Image too large. Max 5MB"})
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExts[ext] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Only jpg and png are allowed"})
	}
	// magic bytes check
	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "We could not read image"})
	}
	defer f.Close()
	header := make([]byte, 512)
	n, _ := f.Read(header)
	header = header[:n]
	if err := ValidateProofFile(file.Filename, file.Size, header); err != nil {
		// reuse proof validator logic for jpg/png magic; but allow only jpg/png
		// ValidateProofFile already checks jpg/png magic, so if fails return 400
		// But for images we want clearer message
		if ext == ".png" && len(header) >= 4 && !(header[0] == 0x89 && header[1] == 'P' && header[2] == 'N' && header[3] == 'G') {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid png file"})
		}
		if (ext == ".jpg" || ext == ".jpeg") && !(len(header) >= 3 && header[0] == 0xFF && header[1] == 0xD8 && header[2] == 0xFF) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid jpg file"})
		}
	}
	dir := filepath.Join("storage", "room_types", fmt.Sprintf("%d", roomTypeID))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not save image"})
	}
	filename := uuid.NewString() + ext
	fullPath := filepath.Join(dir, filename)
	if err := c.SaveFile(file, fullPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not save image"})
	}
	url := "/" + filepath.ToSlash(fullPath)
	// also ensure without leading slash? spec says storage/room_types/{id}/{uuid}.ext
	img, err := h.Repo.Create(roomTypeID, url)
	if err != nil {
		_ = os.Remove(fullPath)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not save image record"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": img})
}

// DeleteImage handles DELETE /api/admin/room-types/:id/images/:imageId or /api/admin/room-images/:id
func (h *RoomImageHandler) DeleteImage(c *fiber.Ctx) error {
	// support both routes
	imageIDStr := c.Params("imageId")
	if imageIDStr == "" {
		imageIDStr = c.Params("image_id")
	}
	if imageIDStr == "" {
		imageIDStr = c.Params("id")
		// if route is /room-types/:id/images/:imageId, then id is roomType, need imageId
		// fallback check query
	}
	// try to parse as image id directly
	imgID, err := parseID(imageIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid image id"})
	}
	img, err := h.Repo.GetByID(imgID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Image not found"})
	}
	// delete file if exists
	if img.URL != "" {
		fsPath := strings.TrimPrefix(img.URL, "/")
		_ = os.Remove(fsPath)
	}
	if err := h.Repo.Delete(imgID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not delete image"})
	}
	return c.JSON(fiber.Map{"message": "Image deleted"})
}

func parseID(s string) (int64, error) {
	var id int64
	_, err := fmt.Sscanf(s, "%d", &id)
	if err != nil {
		return 0, err
	}
	if id <= 0 {
		return 0, fmt.Errorf("invalid id")
	}
	return id, nil
}
