package repo

import "github.com/jmoiron/sqlx"

// RoomRepo stub.
type RoomRepo struct {
	DB *sqlx.DB
}

func NewRoomRepo(db *sqlx.DB) *RoomRepo { return &RoomRepo{DB: db} }
