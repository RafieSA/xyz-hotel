package model

import "time"

// Role enum values.
const (
	RoleOwner        = "owner"
	RoleManager      = "manager"
	RoleReceptionist = "receptionist"
	RoleCustomer     = "customer"
)

// User represents users table.
type User struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name" validate:"required,min=2,max=100"`
	Email     string    `db:"email" json:"email" validate:"required,email"`
	Password  string    `db:"password" json:"-"`
	Role      string    `db:"role" json:"role" validate:"required,oneof=owner manager receptionist customer"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}
