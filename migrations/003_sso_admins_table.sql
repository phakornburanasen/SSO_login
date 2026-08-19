-- ============================================================
-- Migration 003 — แยก "admin role" ออกจาก "env ownership"
--
-- ก่อนหน้านี้: IsAdminUser() เช็คจาก sso_environments.ADUser
--   → bug: user ที่มี env 1 ตัว กลายเป็น admin เห็น env ทั้งหมด
--   → bug: logic "เช็คสิทธิ์ login" ผิดเพราะปนกับการเช็คความเป็นเจ้าของ
--
-- หลัง: ใช้ตาราง sso_admins แยก
--   - admin: ผู้ดูแลระบบ เห็น env ทั้งหมด จัดการทุกอย่างได้
--   - user : ผู้ใช้ทั่วไป เห็นเฉพาะ env ที่ตัวเองเป็น ADUser
-- ============================================================

CREATE TABLE IF NOT EXISTS sso_admins (
    admin_id      SERIAL PRIMARY KEY,
    ad_username   VARCHAR(15) NOT NULL UNIQUE,
    display_name  VARCHAR(100),
    note          TEXT,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS ix_sso_admins_active
    ON sso_admins(ad_username) WHERE is_active = TRUE;

COMMENT ON TABLE sso_admins IS
    'รายชื่อ admin — แยกจาก env ownership ใช้สำหรับกำหนดสิทธิ์ระดับระบบ';

-- seed: admin เริ่มต้น (username = admin)
INSERT INTO sso_admins (ad_username, display_name, note)
VALUES ('admin', 'System Administrator', 'seed user — full access')
ON CONFLICT (ad_username) DO NOTHING;
