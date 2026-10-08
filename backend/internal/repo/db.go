package repo

// Ensure pgx, sqlx, migrate, validator are tracked by go.mod (Fase 1: import only).
import (
	_ "github.com/go-playground/validator/v10"
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)
