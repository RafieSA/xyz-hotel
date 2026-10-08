package seed

import (
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// Run seeds room_types, room_units and users. Idempotent via ON CONFLICT.
func Run(db *sqlx.DB) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// --- room_types ---
	type rt struct {
		Name        string
		Description string
		Capacity    int
		Price       int64
		TotalUnits  int
	}
	roomTypes := []rt{
		{"Standard", "Standard room — comfortable for 2 guests", 2, 350000, 8},
		{"Deluxe", "Deluxe room — spacious with premium amenities", 2, 550000, 5},
		{"Family", "Family room — ideal for families up to 4", 4, 850000, 3},
		{"Suite", "Suite — luxury suite with exclusive facilities", 4, 1250000, 2},
	}

	// id by name after upsert
	ids := map[string]int64{}
	for _, r := range roomTypes {
		var id int64
		// upsert then return id
		q := `
		INSERT INTO room_types (name, description, capacity, price, total_units)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (name) DO UPDATE
		  SET description=EXCLUDED.description,
		      capacity=EXCLUDED.capacity,
		      price=EXCLUDED.price,
		      total_units=EXCLUDED.total_units,
		      updated_at=now()
		RETURNING id`
		if err = tx.QueryRowx(q, r.Name, r.Description, r.Capacity, r.Price, r.TotalUnits).Scan(&id); err != nil {
			return fmt.Errorf("seed room_types %s: %w", r.Name, err)
		}
		ids[r.Name] = id
		slog.Info("seed room_type", "name", r.Name, "id", id)
	}

	// --- room_units ---
	type unit struct {
		Type string
		Code string
	}
	units := []unit{
		{"Standard", "STD-101"}, {"Standard", "STD-102"}, {"Standard", "STD-103"}, {"Standard", "STD-104"},
		{"Standard", "STD-105"}, {"Standard", "STD-106"}, {"Standard", "STD-107"}, {"Standard", "STD-108"},
		{"Deluxe", "DLX-201"}, {"Deluxe", "DLX-202"}, {"Deluxe", "DLX-203"}, {"Deluxe", "DLX-204"}, {"Deluxe", "DLX-205"},
		{"Family", "FAM-301"}, {"Family", "FAM-302"}, {"Family", "FAM-303"},
		{"Suite", "STE-401"}, {"Suite", "STE-402"},
	}
	for _, u := range units {
		rtID, ok := ids[u.Type]
		if !ok {
			return fmt.Errorf("room_type not found: %s", u.Type)
		}
		q := `
		INSERT INTO room_units (room_type_id, code, status)
		VALUES ($1,$2,'available')
		ON CONFLICT (code) DO UPDATE
		  SET room_type_id=EXCLUDED.room_type_id,
		      updated_at=now()
		`
		if _, err = tx.Exec(q, rtID, u.Code); err != nil {
			return fmt.Errorf("seed room_units %s: %w", u.Code, err)
		}
	}

	// --- users with bcrypt ---
	type u struct {
		Name     string
		Email    string
		Password string
		Role     string
	}
	users := []u{
		{"Owner", "owner@xyz-hotel.local", "Owner123!", "owner"},
		{"Manager", "manager@xyz-hotel.local", "Manager123!", "manager"},
		{"Receptionist", "receptionist@xyz-hotel.local", "Receptionist123!", "receptionist"},
		{"Customer", "customer@xyz-hotel.local", "Customer123!", "customer"},
	}
	for _, usr := range users {
		hash, err2 := bcrypt.GenerateFromPassword([]byte(usr.Password), bcrypt.DefaultCost)
		if err2 != nil {
			return fmt.Errorf("bcrypt %s: %w", usr.Email, err2)
		}
		q := `
		INSERT INTO users (name, email, password, role)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (email) DO UPDATE
		  SET name=EXCLUDED.name,
		      password=EXCLUDED.password,
		      role=EXCLUDED.role,
		      updated_at=now()
		`
		if _, err = tx.Exec(q, usr.Name, usr.Email, string(hash), usr.Role); err != nil {
			return fmt.Errorf("seed users %s: %w", usr.Email, err)
		}
		slog.Info("seed user", "email", usr.Email, "role", usr.Role)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	slog.Info("seed completed", "room_types", len(roomTypes), "room_units", len(units), "users", len(users))
	return nil
}
