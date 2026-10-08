package service

import (
	"xyz-hotel/backend/internal/repo"

	"github.com/jmoiron/sqlx"
)

func newSvcReal(db *sqlx.DB) *ReviewService {
	return NewReviewService(repo.NewReviewRepo(db), repo.NewBookingRepo(db), repo.NewRoomRepo(db))
}

// Ensure imports used
var _ = repo.NewReviewRepo
