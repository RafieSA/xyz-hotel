-- 006_addons.sql — addons catalog + booking_addons M:N with price snapshot

CREATE TABLE IF NOT EXISTS addons (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL UNIQUE,
  price BIGINT NOT NULL CHECK (price >= 0),
  description TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO addons (name, price, description) VALUES
  ('Breakfast', 50000, 'Daily breakfast for 2'),
  ('Airport transfer', 150000, 'Airport pickup and drop off'),
  ('Extra bed', 100000, 'Extra bed for additional guest')
ON CONFLICT (name) DO UPDATE SET price = EXCLUDED.price, description = EXCLUDED.description;

CREATE TABLE IF NOT EXISTS booking_addons (
  booking_id BIGINT NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
  addon_id BIGINT NOT NULL REFERENCES addons(id) ON DELETE RESTRICT,
  price BIGINT NOT NULL CHECK (price >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (booking_id, addon_id)
);
CREATE INDEX IF NOT EXISTS idx_booking_addons_booking ON booking_addons(booking_id);
CREATE INDEX IF NOT EXISTS idx_booking_addons_addon ON booking_addons(addon_id);
