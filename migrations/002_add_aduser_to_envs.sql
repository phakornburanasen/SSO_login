-- ============================================================
-- Migration 002 : เพิ่มคอลัมน์ ADUser ใน sso_environments
-- (idempotent — รันซ้ำได้)
-- ============================================================

ALTER TABLE sso_environments
    ADD COLUMN IF NOT EXISTS "ADUser" VARCHAR(15);
