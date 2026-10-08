package repo

import (
	"context"

	"xyz-hotel/backend/internal/model"

	"github.com/jmoiron/sqlx"
)

// UserRepo handles users queries with $1 placeholders.
type UserRepo struct {
	DB *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *UserRepo { return &UserRepo{DB: db} }

// FindByEmail returns user by email or error. ctx is accepted for compatibility with service layer.
func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := r.DB.GetContext(ctx, &u, `SELECT id, name, email, password, role, created_at, updated_at, deleted_at FROM users WHERE email=$1 AND deleted_at IS NULL`, email)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByEmailNoCtx convenience wrapper.
func (r *UserRepo) FindByEmailNoCtx(email string) (*model.User, error) {
	return r.FindByEmail(context.Background(), email)
}

// FindByID returns user by id.
func (r *UserRepo) FindByID(ctx context.Context, id int64) (*model.User, error) {
	var u model.User
	err := r.DB.GetContext(ctx, &u, `SELECT id, name, email, password, role, created_at, updated_at, deleted_at FROM users WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Create inserts a new user. Sets ID on u and returns error.
func (r *UserRepo) Create(ctx context.Context, u *model.User) error {
	var created model.User
	err := r.DB.GetContext(ctx, &created,
		`INSERT INTO users (name, email, password, role) VALUES ($1,$2,$3,$4)
		 RETURNING id, name, email, password, role, created_at, updated_at, deleted_at`,
		u.Name, u.Email, u.Password, u.Role)
	if err != nil {
		return err
	}
	*u = created
	return nil
}

// CreateWithReturn inserts and returns created user (legacy helper).
func (r *UserRepo) CreateWithReturn(u *model.User) (*model.User, error) {
	if err := r.Create(context.Background(), u); err != nil {
		return nil, err
	}
	return u, nil
}
