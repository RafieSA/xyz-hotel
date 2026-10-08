-- 001_init.sql — xyz-hotel Fase 1
-- Tables: users, room_types, room_units, bookings, vouchers, audit_logs

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Enums
DO $$ BEGIN
  CREATE TYPE user_role AS ENUM ('owner','manager','receptionist','customer');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
  CREATE TYPE room_unit_status AS ENUM ('available','occupied','dirty','maintenance');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
  CREATE TYPE booking_status AS ENUM ('pending_payment','waiting_verification','verified','checked_in','checked_out','rejected','expired','cancelled');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

-- users
CREATE TABLE IF NOT EXISTS users (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL,
  role user_role NOT NULL DEFAULT 'customer',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- room_types
CREATE TABLE IF NOT EXISTS room_types (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT '',
  capacity INT NOT NULL CHECK (capacity >= 1),
  price BIGINT NOT NULL CHECK (price >= 0),
  total_units INT NOT NULL DEFAULT 0 CHECK (total_units >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- room_units (physical units)
CREATE TABLE IF NOT EXISTS room_units (
  id BIGSERIAL PRIMARY KEY,
  room_type_id BIGINT NOT NULL REFERENCES room_types(id) ON DELETE RESTRICT,
  code VARCHAR(50) NOT NULL UNIQUE,
  status room_unit_status NOT NULL DEFAULT 'available',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_room_units_room_type ON room_units(room_type_id);
CREATE INDEX IF NOT EXISTS idx_room_units_status ON room_units(status);

-- vouchers
CREATE TABLE IF NOT EXISTS vouchers (
  id BIGSERIAL PRIMARY KEY,
  code VARCHAR(50) NOT NULL UNIQUE,
  discount_percent DOUBLE PRECISION NOT NULL CHECK (discount_percent >= 0 AND discount_percent <= 100),
  min_nights INT NOT NULL DEFAULT 0 CHECK (min_nights >= 0),
  quota INT CHECK (quota IS NULL OR quota >= 0),
  used_count INT NOT NULL DEFAULT 0 CHECK (used_count >= 0),
  expires_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_vouchers_code ON vouchers(code);

-- bookings
CREATE TABLE IF NOT EXISTS bookings (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  room_type_id BIGINT NOT NULL REFERENCES room_types(id) ON DELETE RESTRICT,
  room_unit_id BIGINT REFERENCES room_units(id) ON DELETE SET NULL,
  check_in DATE NOT NULL,
  check_out DATE NOT NULL,
  guests INT NOT NULL CHECK (guests >= 1),
  total_price BIGINT NOT NULL CHECK (total_price >= 0),
  status booking_status NOT NULL DEFAULT 'pending_payment',
  voucher_id BIGINT REFERENCES vouchers(id) ON DELETE SET NULL,
  proof_url TEXT,
  reject_reason TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CHECK (check_out > check_in)
);
CREATE INDEX IF NOT EXISTS idx_bookings_user ON bookings(user_id);
CREATE INDEX IF NOT EXISTS idx_bookings_room_type ON bookings(room_type_id);
CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings(status);
CREATE INDEX IF NOT EXISTS idx_bookings_dates ON bookings(check_in, check_out);
CREATE INDEX IF NOT EXISTS idx_bookings_overlap ON bookings(room_type_id, status, check_in, check_out);

-- audit_logs
CREATE TABLE IF NOT EXISTS audit_logs (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
  action VARCHAR(100) NOT NULL,
  entity VARCHAR(100) NOT NULL DEFAULT '',
  entity_id BIGINT,
  payload JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs(entity, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created ON audit_logs(created_at);
