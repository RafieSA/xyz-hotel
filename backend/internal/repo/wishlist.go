package repo

import (
	"xyz-hotel/backend/internal/model"

	"github.com/jmoiron/sqlx"
)

// WishlistRepo handles wishlists with $1 placeholders.
type WishlistRepo struct {
	DB *sqlx.DB
}

func NewWishlistRepo(db *sqlx.DB) *WishlistRepo { return &WishlistRepo{DB: db} }

// Toggle inserts or deletes wishlist entry. Returns true if added, false if removed.
func (r *WishlistRepo) Toggle(userID, roomTypeID int64) (bool, error) {
	var exists bool
	if err := r.DB.Get(&exists, `SELECT EXISTS(SELECT 1 FROM room_types WHERE id=$1)`, roomTypeID); err != nil {
		return false, err
	}
	if !exists {
		return false, errRoomTypeNotFound
	}
	res, err := r.DB.Exec(`DELETE FROM wishlists WHERE user_id=$1 AND room_type_id=$2`, userID, roomTypeID)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		return false, nil
	}
	_, err = r.DB.Exec(`INSERT INTO wishlists (user_id, room_type_id) VALUES ($1,$2)`, userID, roomTypeID)
	if err != nil {
		return false, err
	}
	return true, nil
}

var errRoomTypeNotFound = errNotFound("room type not found")

type errNotFound string

func (e errNotFound) Error() string { return string(e) }

// IsRoomTypeNotFound checks if error is room type not found.
func IsRoomTypeNotFound(err error) bool {
	_, ok := err.(errNotFound)
	return ok
}

// List returns wishlists for user.
func (r *WishlistRepo) List(userID int64) ([]model.Wishlist, error) {
	var list []model.Wishlist
	err := r.DB.Select(&list, `SELECT id, user_id, room_type_id, created_at FROM wishlists WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.Wishlist{}
	}
	return list, nil
}

// ListEnriched returns wishlists enriched with room_type details for frontend heart display.
func (r *WishlistRepo) ListEnriched(userID int64) ([]map[string]interface{}, error) {
	list, err := r.List(userID)
	if err != nil {
		return nil, err
	}
	var out []map[string]interface{}
	for _, w := range list {
		var rt model.RoomType
		err = r.DB.Get(&rt, `SELECT id, name, description, capacity, price, total_units, created_at, updated_at FROM room_types WHERE id=$1`, w.RoomTypeID)
		if err != nil {
			continue
		}
		var avgRating float64
		var reviewCount int
		_ = r.DB.Get(&avgRating, `SELECT COALESCE(avg_rating,0) FROM room_types WHERE id=$1`, w.RoomTypeID)
		_ = r.DB.Get(&reviewCount, `SELECT COALESCE(review_count,0) FROM room_types WHERE id=$1`, w.RoomTypeID)
		out = append(out, map[string]interface{}{
			"id":           w.ID,
			"user_id":      w.UserID,
			"room_type_id": w.RoomTypeID,
			"created_at":   w.CreatedAt,
			"room_type": map[string]interface{}{
				"id":           rt.ID,
				"name":         rt.Name,
				"description":  rt.Description,
				"capacity":     rt.Capacity,
				"price":        rt.Price,
				"total_units":  rt.TotalUnits,
				"avg_rating":   avgRating,
				"review_count": reviewCount,
			},
		})
	}
	if out == nil {
		out = []map[string]interface{}{}
	}
	return out, nil
}

// Delete removes wishlist entry for user+room_type.
func (r *WishlistRepo) Delete(userID, roomTypeID int64) (bool, error) {
	res, err := r.DB.Exec(`DELETE FROM wishlists WHERE user_id=$1 AND room_type_id=$2`, userID, roomTypeID)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// Exists checks if wishlist exists.
func (r *WishlistRepo) Exists(userID, roomTypeID int64) (bool, error) {
	var exists bool
	err := r.DB.Get(&exists, `SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id=$1 AND room_type_id=$2)`, userID, roomTypeID)
	return exists, err
}
