-- 007_loyalty.sql — loyalty points on users + ledger table loyalty_tx

ALTER TABLE users ADD COLUMN IF NOT EXISTS loyalty_points INT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS loyalty_tx (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  booking_id BIGINT REFERENCES bookings(id) ON DELETE SET NULL,
  points INT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_loyalty_tx_user ON loyalty_tx(user_id);
CREATE INDEX IF NOT EXISTS idx_loyalty_tx_booking ON loyalty_tx(booking_id);
