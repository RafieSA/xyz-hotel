package main

import (
	"log/slog"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/jackc/pgx/v5/stdlib"

	"xyz-hotel/backend/seed"
)

func main() {
	_ = godotenv.Load()
	// also try backend/.env when run from repo root
	_ = godotenv.Load("backend/.env")

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://rafiesafarazaribowo@localhost:5432/xyz_hotel?sslmode=disable"
		slog.Warn("DATABASE_URL empty, using default", "dsn", dsn)
	}

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		slog.Error("connect db failed", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := seed.Run(db); err != nil {
		slog.Error("seed failed", "err", err)
		os.Exit(1)
	}
	slog.Info("seed done")
}
