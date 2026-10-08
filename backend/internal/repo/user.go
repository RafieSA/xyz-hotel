package repo

import "github.com/jmoiron/sqlx"

// UserRepo stub.
type UserRepo struct {
	DB *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *UserRepo { return &UserRepo{DB: db} }
