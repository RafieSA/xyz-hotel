package repo

import (
	"context"
	"database/sql"
	"fmt"

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

// FindAvailableUnitForType returns id of first available unit for room type. Uses FOR UPDATE.
func (r *RoomRepo) FindAvailableUnitForType(ctx context.Context, roomTypeID int64) (int64, error) {
	var id int64
	err := r.DB.GetContext(ctx, &id, `SELECT id FROM room_units WHERE room_type_id=$1 AND status='available' LIMIT 1 FOR UPDATE`, roomTypeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("no available units for room type %d", roomTypeID)
		}
		return 0, err
	}
	return id, nil
}

// FindAvailableUnitForTypeTx finds available unit inside a transaction with row lock.
func (r *RoomRepo) FindAvailableUnitForTypeTx(ctx context.Context, tx *sqlx.Tx, roomTypeID int64) (int64, error) {
	var id int64
	err := tx.GetContext(ctx, &id, `SELECT id FROM room_units WHERE room_type_id=$1 AND status='available' LIMIT 1 FOR UPDATE`, roomTypeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("no available units for room type %d", roomTypeID)
		}
		return 0, err
	}
	return id, nil
}

// GetUnitByID returns room_unit by id.
func (r *RoomRepo) GetUnitByID(ctx context.Context, unitID int64) (*model.RoomUnit, error) {
	var u model.RoomUnit
	err := r.DB.GetContext(ctx, &u, `SELECT id, room_type_id, code, status, created_at, updated_at, deleted_at FROM room_units WHERE id=$1`, unitID)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUnitByIDTx fetches unit inside transaction with FOR UPDATE.
func (r *RoomRepo) GetUnitByIDTx(ctx context.Context, tx *sqlx.Tx, unitID int64) (*model.RoomUnit, error) {
	var u model.RoomUnit
	err := tx.GetContext(ctx, &u, `SELECT id, room_type_id, code, status, created_at, updated_at, deleted_at FROM room_units WHERE id=$1 FOR UPDATE`, unitID)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// UpdateStatus updates room_unit status (outside transaction).
func (r *RoomRepo) UpdateStatus(ctx context.Context, unitID int64, status string) error {
	res, err := r.DB.ExecContext(ctx, `UPDATE room_units SET status=$1, updated_at=now() WHERE id=$2`, status, unitID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// UpdateStatusTx updates status inside transaction.
func (r *RoomRepo) UpdateStatusTx(ctx context.Context, tx *sqlx.Tx, unitID int64, status string) error {
	res, err := tx.ExecContext(ctx, `UPDATE room_units SET status=$1, updated_at=now() WHERE id=$2`, status, unitID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ListUnitsByType returns all units for a room type.
func (r *RoomRepo) ListUnitsByType(ctx context.Context, roomTypeID int64) ([]model.RoomUnit, error) {
	var out []model.RoomUnit
	err := r.DB.SelectContext(ctx, &out, `SELECT id, room_type_id, code, status, created_at, updated_at, deleted_at FROM room_units WHERE room_type_id=$1 ORDER BY code ASC`, roomTypeID)
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.RoomUnit{}
	}
	return out, nil
}

// ListAllUnits returns all room units.
func (r *RoomRepo) ListAllUnits(ctx context.Context) ([]model.RoomUnit, error) {
	var out []model.RoomUnit
	err := r.DB.SelectContext(ctx, &out, `SELECT id, room_type_id, code, status, created_at, updated_at, deleted_at FROM room_units ORDER BY room_type_id, code`)
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.RoomUnit{}
	}
	return out, nil
}
