package repo

import (
	"fmt"

	"xyz-hotel/backend/internal/model"

	"github.com/jmoiron/sqlx"
)

// AuditRepo handles audit_logs queries.
type AuditRepo struct {
	DB *sqlx.DB
}

func NewAuditRepo(db *sqlx.DB) *AuditRepo { return &AuditRepo{DB: db} }

// List returns audit logs ordered by created_at desc with pagination.
// limit clamped 1..100, offset >=0. Optional filters: entity, action.
func (r *AuditRepo) List(limit, offset int, entity, action string) ([]model.AuditLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	query := `SELECT id, user_id, action, entity, entity_id, payload, created_at FROM audit_logs WHERE 1=1`
	args := []interface{}{}
	idx := 1
	if entity != "" {
		query += fmt.Sprintf(` AND entity=$%d`, idx)
		args = append(args, entity)
		idx++
	}
	if action != "" {
		query += fmt.Sprintf(` AND action=$%d`, idx)
		args = append(args, action)
		idx++
	}
	query += fmt.Sprintf(` ORDER BY created_at DESC, id DESC LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, limit, offset)

	var out []model.AuditLog
	if err := r.DB.Select(&out, query, args...); err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.AuditLog{}
	}
	return out, nil
}

// Count returns total count for filters.
func (r *AuditRepo) Count(entity, action string) (int, error) {
	query := `SELECT COUNT(*) FROM audit_logs WHERE 1=1`
	args := []interface{}{}
	idx := 1
	if entity != "" {
		query += fmt.Sprintf(` AND entity=$%d`, idx)
		args = append(args, entity)
		idx++
	}
	if action != "" {
		query += fmt.Sprintf(` AND action=$%d`, idx)
		args = append(args, action)
		idx++
	}
	var cnt int
	if err := r.DB.Get(&cnt, query, args...); err != nil {
		return 0, err
	}
	return cnt, nil
}
