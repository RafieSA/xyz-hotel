-- 004_room_soft.sql — soft delete for room_types
ALTER TABLE room_types ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_room_types_deleted ON room_types(deleted_at);
