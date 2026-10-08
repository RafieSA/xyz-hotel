# xyz-hotel backend

Go 1.26 + Fiber v2 + pgx/sqlx

## Prasyarat
- Go 1.26.3
- Postgres 18.4 (local, tanpa Docker)

## Cara jalan

```bash
cp .env.example .env
# edit DATABASE_URL & JWT_SECRET

# install deps
go mod tidy

# migrasi (butuh golang-migrate CLI)
# install: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
migrate -path migrations -database "$DATABASE_URL" up
# atau via Makefile (butuh DATABASE_URL di env)
make migrate

# jalan
go run ./cmd/server
# atau
make run
```

Health check:
```
GET http://localhost:8080/health
GET http://localhost:8080/api/health
→ {"status":"ok","service":"xyz-hotel"}
```

## Struktur
```
cmd/server/main.go
internal/handler/   # HTTP
internal/service/   # business (availability FOR UPDATE)
internal/middleware/# JWT + RBAC
internal/model/     # structs
internal/repo/      # sqlx
migrations/001_init.sql
```

## Makefile
- `make run` - jalan server
- `make migrate` - migrate up
- `make tidy` - go mod tidy
- `make vet` - go vet ./...
