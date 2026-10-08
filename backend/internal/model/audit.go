package model

import "time"

// AuditLog represents audit_logs table.
type AuditLog struct {
	ID        int64     `db:"id" json:"id"`
	UserID    *int64    `db:"user_id" json:"user_id,omitempty"`
	Action    string    `db:"action" json:"action" validate:"required"`
	Entity    string    `db:"entity" json:"entity"`
	EntityID  *int64    `db:"entity_id" json:"entity_id,omitempty"`
	Payload   *string   `db:"payload" json:"payload,omitempty"` // JSON string
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
