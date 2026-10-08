package service

import "github.com/jmoiron/sqlx"

// AvailabilityService handles availability checks with ACID guarantees.
type AvailabilityService struct {
	DB *sqlx.DB
}

func NewAvailabilityService(db *sqlx.DB) *AvailabilityService {
	return &AvailabilityService{DB: db}
}

// CheckAvailability returns remaining units for a room type and date range.
// Must be called inside a transaction with FOR UPDATE to prevent race conditions.
//
// Example (to be used in booking creation transaction):
//
//   tx, _ := s.DB.BeginTxx(ctx, nil)
//   // SELECT COUNT(*) FROM bookings WHERE room_type_id=$1 AND status IN ('verified','checked_in') AND check_in < $3 AND check_out > $2 FOR UPDATE
//   // compare with room_types.total_units
//   // if occupied >= totalUnits -> reject
//   // else INSERT booking
//   // tx.Commit()
func (s *AvailabilityService) CheckAvailability(roomTypeID int64, checkIn, checkOut string) (int, error) {
	// stub Fase 1
	return 0, nil
}
