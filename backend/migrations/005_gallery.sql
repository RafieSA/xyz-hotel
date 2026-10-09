-- 005_gallery.sql — room_type_images gallery (max 5 per room type enforced in service/handler)

CREATE TABLE IF NOT EXISTS room_type_images (
  id BIGSERIAL PRIMARY KEY,
  room_type_id BIGINT NOT NULL REFERENCES room_types(id) ON DELETE CASCADE,
  url TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_room_type_images_room_type ON room_type_images(room_type_id);
