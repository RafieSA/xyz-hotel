package repo

import (
	"context"
	"database/sql"

	"xyz-hotel/backend/internal/model"

	"github.com/jmoiron/sqlx"
)

// VoucherRepo handles vouchers queries with $1 placeholders (SRP, $1 safe).
type VoucherRepo struct {
	DB *sqlx.DB
}

func NewVoucherRepo(db *sqlx.DB) *VoucherRepo { return &VoucherRepo{DB: db} }

// FindByCode returns voucher by code or sql.ErrNoRows.
func (r *VoucherRepo) FindByCode(ctx context.Context, code string) (*model.Voucher, error) {
	var v model.Voucher
	err := r.DB.GetContext(ctx, &v,
		`SELECT id, code, discount_percent, min_nights, quota, used_count, expires_at, created_at, updated_at
		 FROM vouchers WHERE code=$1`, code)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// FindByCodeForUpdate fetches voucher with row lock inside a transaction (FOR UPDATE).
func (r *VoucherRepo) FindByCodeForUpdate(tx *sqlx.Tx, code string) (*model.Voucher, error) {
	var v model.Voucher
	err := tx.Get(&v,
		`SELECT id, code, discount_percent, min_nights, quota, used_count, expires_at, created_at, updated_at
		 FROM vouchers WHERE code=$1 FOR UPDATE`, code)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// FindByID returns voucher by id.
func (r *VoucherRepo) FindByID(ctx context.Context, id int64) (*model.Voucher, error) {
	var v model.Voucher
	err := r.DB.GetContext(ctx, &v,
		`SELECT id, code, discount_percent, min_nights, quota, used_count, expires_at, created_at, updated_at
		 FROM vouchers WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// Create inserts a voucher and returns the created row with $1 placeholders.
func (r *VoucherRepo) Create(ctx context.Context, v *model.Voucher) (*model.Voucher, error) {
	var created model.Voucher
	err := r.DB.GetContext(ctx, &created,
		`INSERT INTO vouchers (code, discount_percent, min_nights, quota, expires_at)
		 VALUES ($1,$2,$3,$4,$5)
		 RETURNING id, code, discount_percent, min_nights, quota, used_count, expires_at, created_at, updated_at`,
		v.Code, v.Discount, v.MinNights, v.Quota, v.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

// IncrementUsedCount increments used_count outside a transaction.
func (r *VoucherRepo) IncrementUsedCount(ctx context.Context, id int64) error {
	res, err := r.DB.ExecContext(ctx,
		`UPDATE vouchers SET used_count = used_count + 1, updated_at = now() WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// IncrementUsedCountTx increments used_count inside a transaction (caller holds FOR UPDATE lock).
func (r *VoucherRepo) IncrementUsedCountTx(tx *sqlx.Tx, id int64) error {
	res, err := tx.Exec(`UPDATE vouchers SET used_count = used_count + 1, updated_at = now() WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// List returns all vouchers ordered by created_at desc.
func (r *VoucherRepo) List(ctx context.Context) ([]model.Voucher, error) {
	var out []model.Voucher
	err := r.DB.SelectContext(ctx, &out,
		`SELECT id, code, discount_percent, min_nights, quota, used_count, expires_at, created_at, updated_at
		 FROM vouchers ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.Voucher{}
	}
	return out, nil
}
