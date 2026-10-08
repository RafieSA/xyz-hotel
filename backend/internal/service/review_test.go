package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestCreateReview_RatingBounds(t *testing.T) {
	db, _, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer db.Close()
	svc := newSvcReal(sqlxDB)
	ctx := context.Background()
	_, err := svc.CreateReview(ctx, 1, 1, 0, "ok")
	if err != ErrRatingBounds {
		t.Fatalf("expected ErrRatingBounds got %v", err)
	}
	_, err = svc.CreateReview(ctx, 1, 1, 6, "ok")
	if err != ErrRatingBounds {
		t.Fatalf("expected ErrRatingBounds for 6 got %v", err)
	}
	_, err = svc.CreateReview(ctx, 1, 1, 3, string(make([]byte, 501)))
	if err != ErrCommentTooLong {
		t.Fatalf("expected ErrCommentTooLong got %v", err)
	}
}

func TestCreateReview_NotCheckedOut(t *testing.T) {
	db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer db.Close()
	svc := newSvcReal(sqlxDB)
	ctx := context.Background()
	now := time.Now()
	mock.ExpectQuery(`SELECT.*bookings`).WithArgs(int64(10)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "user_id", "room_type_id", "room_unit_id", "check_in", "check_out", "guests", "total_price", "status", "voucher_id", "proof_url", "reject_reason", "created_at", "updated_at"}).
			AddRow(int64(10), int64(1), int64(2), nil, now, now.Add(24*time.Hour), 2, int64(1000), "verified", nil, nil, nil, now, now),
	)
	_, err := svc.CreateReview(ctx, 1, 10, 5, "nice")
	if err != ErrNotCheckedOut {
		t.Fatalf("expected ErrNotCheckedOut got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateReview_Forbidden(t *testing.T) {
	db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer db.Close()
	svc := newSvcReal(sqlxDB)
	ctx := context.Background()
	now := time.Now()
	mock.ExpectQuery(`SELECT.*bookings`).WithArgs(int64(11)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "user_id", "room_type_id", "room_unit_id", "check_in", "check_out", "guests", "total_price", "status", "voucher_id", "proof_url", "reject_reason", "created_at", "updated_at"}).
			AddRow(int64(11), int64(99), int64(2), nil, now, now.Add(24*time.Hour), 2, int64(1000), "checked_out", nil, nil, nil, now, now),
	)
	_, err := svc.CreateReview(ctx, 1, 11, 5, "nice")
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateReview_AlreadyReviewed(t *testing.T) {
	db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer db.Close()
	svc := newSvcReal(sqlxDB)
	ctx := context.Background()
	now := time.Now()
	mock.ExpectQuery(`SELECT.*bookings`).WithArgs(int64(12)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "user_id", "room_type_id", "room_unit_id", "check_in", "check_out", "guests", "total_price", "status", "voucher_id", "proof_url", "reject_reason", "created_at", "updated_at"}).
			AddRow(int64(12), int64(1), int64(2), nil, now, now.Add(24*time.Hour), 2, int64(1000), "checked_out", nil, nil, nil, now, now),
	)
	mock.ExpectQuery(`SELECT \* FROM reviews`).WithArgs(int64(12)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "booking_id", "user_id", "room_type_id", "rating", "comment", "created_at"}).
			AddRow(int64(1), int64(12), int64(1), int64(2), 5, "old", now),
	)
	_, err := svc.CreateReview(ctx, 1, 12, 5, "again")
	if err != ErrAlreadyReviewed {
		t.Fatalf("expected ErrAlreadyReviewed got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateReview_Success(t *testing.T) {
	db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer db.Close()
	svc := newSvcReal(sqlxDB)
	ctx := context.Background()
	now := time.Now()
	mock.ExpectQuery(`SELECT.*bookings`).WithArgs(int64(13)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "user_id", "room_type_id", "room_unit_id", "check_in", "check_out", "guests", "total_price", "status", "voucher_id", "proof_url", "reject_reason", "created_at", "updated_at"}).
			AddRow(int64(13), int64(1), int64(2), nil, now, now.Add(24*time.Hour), 2, int64(1000), "checked_out", nil, nil, nil, now, now),
	)
	mock.ExpectQuery(`SELECT \* FROM reviews`).WithArgs(int64(13)).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`INSERT INTO reviews`).WithArgs(int64(13), int64(1), int64(2), 5, "great").WillReturnRows(
		sqlmock.NewRows([]string{"id", "created_at"}).AddRow(int64(100), now),
	)
	mock.ExpectExec(`UPDATE room_types SET`).WithArgs(int64(2)).WillReturnResult(sqlmock.NewResult(1, 1))
	rv, err := svc.CreateReview(ctx, 1, 13, 5, "great")
	if err != nil {
		t.Fatalf("unexpected err %v", err)
	}
	if rv.ID != 100 {
		t.Fatalf("expected id 100 got %d", rv.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
