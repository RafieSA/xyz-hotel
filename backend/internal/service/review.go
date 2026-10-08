package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"xyz-hotel/backend/internal/model"
	"xyz-hotel/backend/internal/repo"
)

var (
	ErrRatingBounds    = errors.New("Rating must be 1 to 5. Please choose a rating from 1 to 5")
	ErrCommentTooLong  = errors.New("Comment is too long. Keep it to 500 characters or fewer")
	ErrNotCheckedOut   = errors.New("You can only review stays you have checked out from. Complete your stay first")
	ErrForbidden       = errors.New("You can only review your own bookings")
	ErrAlreadyReviewed = errors.New("You already reviewed this booking")
)

type ReviewService struct {
	ReviewRepo  *repo.ReviewRepo
	BookingRepo *repo.BookingRepo
	RoomRepo    *repo.RoomRepo // not strictly needed but for validation
}

func NewReviewService(rr *repo.ReviewRepo, br *repo.BookingRepo, roomRepo *repo.RoomRepo) *ReviewService {
	return &ReviewService{ReviewRepo: rr, BookingRepo: br, RoomRepo: roomRepo}
}

func (s *ReviewService) CreateReview(ctx context.Context, userID, bookingID int64, rating int, comment string) (*model.Review, error) {
	if rating < model.ReviewRatingMin || rating > model.ReviewRatingMax {
		return nil, ErrRatingBounds
	}
	if len(comment) > model.MaxCommentLen {
		return nil, ErrCommentTooLong
	}
	comment = strings.TrimSpace(comment)

	booking, err := s.BookingRepo.GetByID(bookingID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("Booking not found")
		}
		return nil, err
	}
	if booking.UserID != userID {
		return nil, ErrForbidden
	}
	if booking.Status != model.BookingCheckedOut {
		return nil, ErrNotCheckedOut
	}
	// check one per booking
	existing, err := s.ReviewRepo.FindByBookingID(ctx, bookingID)
	if err == nil && existing != nil {
		return nil, ErrAlreadyReviewed
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	rv := &model.Review{
		BookingID:  bookingID,
		UserID:     userID,
		RoomTypeID: booking.RoomTypeID,
		Rating:     rating,
		Comment:    comment,
	}
	created, err := s.ReviewRepo.Create(ctx, rv)
	if err != nil {
		// unique violation on booking_id -> already reviewed
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, ErrAlreadyReviewed
		}
		return nil, err
	}
	// update stats
	_ = s.ReviewRepo.UpdateRoomTypeStats(ctx, booking.RoomTypeID)
	return created, nil
}

func (s *ReviewService) ListReviews(ctx context.Context, roomTypeID int64) ([]model.Review, error) {
	return s.ReviewRepo.ListByRoomType(ctx, roomTypeID)
}
