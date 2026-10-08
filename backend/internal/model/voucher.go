package model

import "time"

// Voucher represents vouchers table.
type Voucher struct {
	ID        int64      `db:"id" json:"id"`
	Code      string     `db:"code" json:"code" validate:"required"`
	Discount  float64    `db:"discount_percent" json:"discount_percent" validate:"gte=0,lte=100"`
	MinNights int        `db:"min_nights" json:"min_nights" validate:"gte=0"`
	Quota     *int       `db:"quota" json:"quota,omitempty"`
	UsedCount int        `db:"used_count" json:"used_count" validate:"gte=0"`
	ExpiresAt *time.Time `db:"expires_at" json:"expires_at,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
}
