package repo

import (
	"xyz-hotel/backend/internal/model"

	"github.com/jmoiron/sqlx"
)

// AddonRepo handles addons and booking_addons.
type AddonRepo struct {
	DB *sqlx.DB
}

func NewAddonRepo(db *sqlx.DB) *AddonRepo { return &AddonRepo{DB: db} }

// List returns all addons ordered by id.
func (r *AddonRepo) List() ([]model.Addon, error) {
	var out []model.Addon
	err := r.DB.Select(&out, `SELECT id, name, price, description, created_at FROM addons ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.Addon{}
	}
	return out, nil
}

// GetByID fetches addon by id with $1.
func (r *AddonRepo) GetByID(id int64) (*model.Addon, error) {
	var a model.Addon
	err := r.DB.Get(&a, `SELECT id, name, price, description, created_at FROM addons WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// ListByIDs fetches addons for given ids.
func (r *AddonRepo) ListByIDs(ids []int64) ([]model.Addon, error) {
	if len(ids) == 0 {
		return []model.Addon{}, nil
	}
	query, args, err := sqlx.In(`SELECT id, name, price, description, created_at FROM addons WHERE id IN (?)`, ids)
	if err != nil {
		return nil, err
	}
	query = r.DB.Rebind(query)
	var out []model.Addon
	if err := r.DB.Select(&out, query, args...); err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.Addon{}
	}
	return out, nil
}

// FindForBooking returns booking_addons for a booking with joined addon name.
func (r *AddonRepo) FindForBooking(bookingID int64) ([]model.BookingAddon, error) {
	var out []model.BookingAddon
	err := r.DB.Select(&out,
		`SELECT ba.booking_id, ba.addon_id, ba.price, ba.created_at, a.name AS addon_name
		 FROM booking_addons ba JOIN addons a ON a.id = ba.addon_id WHERE ba.booking_id=$1 ORDER BY ba.created_at ASC`, bookingID)
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.BookingAddon{}
	}
	return out, nil
}

// AddToBooking inserts booking_addons rows and updates booking total_price atomically.
// Caller should handle transaction; this helper inserts inside tx.
func (r *AddonRepo) AddToBookingTx(tx *sqlx.Tx, bookingID int64, addons []model.Addon) error {
	for _, a := range addons {
		if _, err := tx.Exec(`INSERT INTO booking_addons (booking_id, addon_id, price) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, bookingID, a.ID, a.Price); err != nil {
			return err
		}
	}
	return nil
}
