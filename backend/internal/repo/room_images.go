package repo

import (
	"xyz-hotel/backend/internal/model"

	"github.com/jmoiron/sqlx"
)

// RoomImageRepo handles room_type_images with $1 placeholders.
type RoomImageRepo struct {
	DB *sqlx.DB
}

func NewRoomImageRepo(db *sqlx.DB) *RoomImageRepo { return &RoomImageRepo{DB: db} }

// Create inserts a room_type_image and returns it.
func (r *RoomImageRepo) Create(roomTypeID int64, url string) (*model.RoomTypeImage, error) {
	var img model.RoomTypeImage
	err := r.DB.Get(&img,
		`INSERT INTO room_type_images (room_type_id, url) VALUES ($1,$2)
		 RETURNING id, room_type_id, url, created_at`, roomTypeID, url)
	if err != nil {
		return nil, err
	}
	return &img, nil
}

// List returns images for a room type ordered by created_at.
func (r *RoomImageRepo) List(roomTypeID int64) ([]model.RoomTypeImage, error) {
	var out []model.RoomTypeImage
	err := r.DB.Select(&out,
		`SELECT id, room_type_id, url, created_at FROM room_type_images WHERE room_type_id=$1 ORDER BY created_at ASC`, roomTypeID)
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.RoomTypeImage{}
	}
	return out, nil
}

// Delete removes an image by id and returns deleted row if found.
func (r *RoomImageRepo) Delete(id int64) error {
	_, err := r.DB.Exec(`DELETE FROM room_type_images WHERE id=$1`, id)
	return err
}

// Count returns number of images for a room type.
func (r *RoomImageRepo) Count(roomTypeID int64) (int, error) {
	var c int
	err := r.DB.Get(&c, `SELECT COUNT(*) FROM room_type_images WHERE room_type_id=$1`, roomTypeID)
	return c, err
}

// GetByID fetches single image by id.
func (r *RoomImageRepo) GetByID(id int64) (*model.RoomTypeImage, error) {
	var img model.RoomTypeImage
	err := r.DB.Get(&img, `SELECT id, room_type_id, url, created_at FROM room_type_images WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	return &img, nil
}
