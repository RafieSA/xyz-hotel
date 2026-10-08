package repo

import (
	"log/slog"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/jmoiron/sqlx"
)

// Connect opens a sqlx DB pool to Postgres via pgx stdlib.
// dsn example: postgres://postgres:postgres@localhost:5432/xyz_hotel?sslmode=disable
func Connect(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	slog.Info("db connected", "dsn", dsn)
	return db, nil
}
