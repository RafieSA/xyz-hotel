package repo

import (
	"context"
	"database/sql"

	"xyz-hotel/backend/internal/model"

	"github.com/jmoiron/sqlx"
)

type ReviewRepo struct {
	DB *sqlx.DB
}

func NewReviewRepo(db *sqlx.DB) *ReviewRepo { return &ReviewRepo{DB: db} }

func (r *ReviewRepo) Create(ctx context.Context, rv *model.Review) (*model.Review, error) {
	q := `INSERT INTO reviews (booking_id, user_id, room_type_id, rating, comment)
	      VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at`
	err := r.DB.QueryRowxContext(ctx, q, rv.BookingID, rv.UserID, rv.RoomTypeID, rv.Rating, rv.Comment).Scan(&rv.ID, &rv.CreatedAt)
	if err != nil {
		return nil, err
	}
	return rv, nil
}

func (r *ReviewRepo) FindByBookingID(ctx context.Context, bookingID int64) (*model.Review, error) {
	var rv model.Review
	err := r.DB.GetContext(ctx, &rv, `SELECT * FROM reviews WHERE booking_id=$1`, bookingID)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

func (r *ReviewRepo) ListByRoomType(ctx context.Context, roomTypeID int64) ([]model.Review, error) {
	var list []model.Review
	err := r.DB.SelectContext(ctx, &list, `SELECT * FROM reviews WHERE room_type_id=$1 ORDER BY created_at DESC`, roomTypeID)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.Review{}
	}
	return list, nil
}

func (r *ReviewRepo) UpdateRoomTypeStats(ctx context.Context, roomTypeID int64) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE room_types SET
		  avg_rating = COALESCE((SELECT AVG(rating)::double precision FROM reviews WHERE room_type_id=$1),0),
		  review_count = (SELECT COUNT(*) FROM reviews WHERE room_type_id=$1),
		  updated_at = now()
		WHERE id=$1`, roomTypeID)
	return err
}

func (r *ReviewRepo) GetAverageByRoomType(ctx context.Context, roomTypeID int64) (float64, int, error) {
	var avg sql.NullFloat64
	var cnt int
	row := r.DB.QueryRowxContext(ctx, `SELECT COALESCE(AVG(rating)::double precision,0), COUNT(*) FROM reviews WHERE room_type_id=$1`, roomTypeID)
	if err := row.Scan(&avg, &cnt); err != nil {
		return 0, 0, err
	}
	if !avg.Valid {
		return 0, cnt, nil
	}
	return avg.Float64, cnt, nil
}
