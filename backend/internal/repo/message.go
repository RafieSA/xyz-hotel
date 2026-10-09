package repo

import (
	"xyz-hotel/backend/internal/model"

	"github.com/jmoiron/sqlx"
)

// MessageRepo handles messages queries with $1 placeholders.
type MessageRepo struct {
	DB *sqlx.DB
}

func NewMessageRepo(db *sqlx.DB) *MessageRepo { return &MessageRepo{DB: db} }

// Create inserts a message.
func (r *MessageRepo) Create(m *model.Message) (*model.Message, error) {
	q := `INSERT INTO messages (user_id, booking_id, message, is_admin) VALUES ($1,$2,$3,$4) RETURNING id, created_at`
	err := r.DB.QueryRowx(q, m.UserID, m.BookingID, m.Message, m.IsAdmin).Scan(&m.ID, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// ListByUser returns messages for a user ordered by created_at ASC.
func (r *MessageRepo) ListByUser(userID int64) ([]model.Message, error) {
	var list []model.Message
	err := r.DB.Select(&list, `SELECT * FROM messages WHERE user_id=$1 ORDER BY created_at ASC`, userID)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.Message{}
	}
	return list, nil
}

// ListForAdmin returns all messages, optionally filtered by user_id, ordered by created_at ASC.
func (r *MessageRepo) ListForAdmin(userID *int64) ([]model.Message, error) {
	var list []model.Message
	var err error
	if userID != nil {
		err = r.DB.Select(&list, `SELECT * FROM messages WHERE user_id=$1 ORDER BY created_at ASC`, *userID)
	} else {
		err = r.DB.Select(&list, `SELECT * FROM messages ORDER BY created_at ASC LIMIT 500`)
	}
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.Message{}
	}
	return list, nil
}

// ListByBooking returns messages for a booking.
func (r *MessageRepo) ListByBooking(bookingID int64) ([]model.Message, error) {
	var list []model.Message
	err := r.DB.Select(&list, `SELECT * FROM messages WHERE booking_id=$1 ORDER BY created_at ASC`, bookingID)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.Message{}
	}
	return list, nil
}

// GetByID fetches a message by id.
func (r *MessageRepo) GetByID(id int64) (*model.Message, error) {
	var m model.Message
	err := r.DB.Get(&m, `SELECT * FROM messages WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// GetConversations returns per-user conversation summary for admin inbox.
func (r *MessageRepo) GetConversations() ([]model.Conversation, error) {
	q := `
		SELECT m.user_id, u.name AS user_name, u.email AS user_email,
		       (SELECT message FROM messages WHERE user_id=m.user_id ORDER BY created_at DESC LIMIT 1) AS last_message,
		       MAX(m.created_at) AS last_at,
		       COUNT(*)::int AS count
		FROM messages m
		JOIN users u ON u.id=m.user_id
		GROUP BY m.user_id, u.name, u.email
		ORDER BY last_at DESC`
	var list []model.Conversation
	err := r.DB.Select(&list, q)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.Conversation{}
	}
	return list, nil
}
