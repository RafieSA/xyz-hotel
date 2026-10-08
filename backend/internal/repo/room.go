package repo

import (
	"xyz-hotel/backend/internal/model"

	"github.com/jmoiron/sqlx"
)

// RoomRepo handles room_types and room_units queries with parameterized $1.
type RoomRepo struct {
	DB *sqlx.DB
}

func NewRoomRepo(db *sqlx.DB) *RoomRepo { return &RoomRepo{DB: db} }

// GetRoomTypeByID returns room_type by id or error.
func (r *RoomRepo) GetRoomTypeByID(id int64) (*model.RoomType, error) {
	var rt model.RoomType
	err := r.DB.Get(&rt, `SELECT id, name, description, capacity, price, total_units, created_at, updated_at FROM room_types WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

// GetRoomTypeByIDTx fetches room_type inside a transaction with FOR UPDATE lock on room_types row.
func (r *RoomRepo) GetRoomTypeByIDTx(tx *sqlx.Tx, id int64) (*model.RoomType, error) {
	var rt model.RoomType
	err := tx.Get(&rt, `SELECT id, name, description, capacity, price, total_units, created_at, updated_at FROM room_types WHERE id=$1 FOR UPDATE`, id)
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

// ListRoomTypes returns all room types.
func (r *RoomRepo) ListRoomTypes() ([]model.RoomType, error) {
	var out []model.RoomType
	err := r.DB.Select(&out, `SELECT id, name, description, capacity, price, total_units, created_at, updated_at FROM room_types ORDER BY price ASC`)
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.RoomType{}
	}
	return out, nil
}

// CountUnits returns total_units for a room type (convenience).
func (r *RoomRepo) CountUnits(roomTypeID int64) (int, error) {
	var total int
	err := r.DB.Get(&total, `SELECT total_units FROM room_types WHERE id=$1`, roomTypeID)
	return total, err
}
