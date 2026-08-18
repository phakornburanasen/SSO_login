-- ============================================================
-- SSO Permission Controller - PostgreSQL schema
-- ตรวจสิทธิ์การเข้าระบบจาก (IP + Base URL + AD Username)
-- ============================================================

-- 1) Applications : ระบบที่ต้องคุม เช่น HelpDesk, ERP, CRM
CREATE TABLE IF NOT EXISTS sso_applications (
    app_id          SERIAL PRIMARY KEY,
    app_code        VARCHAR(50)  NOT NULL UNIQUE,         -- 'HELPDESK'
    app_name        VARCHAR(200) NOT NULL,                -- 'HelpDesk System'
    description     TEXT,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ
);

-- 2) Environments : แยก production / test / uat ของแต่ละ app
--    base_url = "http://10.0.32.71/HelpDesk/"  (เก็บ http:// + ip + base)
CREATE TABLE IF NOT EXISTS sso_environments (
    env_id          SERIAL PRIMARY KEY,
    app_id          INT          NOT NULL REFERENCES sso_applications(app_id) ON DELETE CASCADE,
    env_code        VARCHAR(20)  NOT NULL,                -- 'PROD' | 'TEST' | 'UAT'
    env_name        VARCHAR(100) NOT NULL,                -- 'Production'
    base_url        VARCHAR(255) NOT NULL,                -- 'http://10.0.32.71/HelpDesk/'
    host_ip         INET         NULL,                    -- '10.0.32.71'  (แยกไว้ query/lookup)
    base_path       VARCHAR(255) NULL,                    -- '/HelpDesk/'
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ,
    CONSTRAINT uq_sso_env_app_code UNIQUE (app_id, env_code)
);

-- 3) Allowed IPs : IP / CIDR ที่อนุญาตให้เข้า env นั้น
--    ใช้ INET เพื่อรองรับทั้ง IP เดี่ยว (10.0.32.71) และ CIDR (10.0.32.0/24)
CREATE TABLE IF NOT EXISTS sso_allowed_ips (
    allowed_ip_id   SERIAL PRIMARY KEY,
    env_id          INT          NOT NULL REFERENCES sso_environments(env_id) ON DELETE CASCADE,
    ip_cidr         INET         NOT NULL,                -- '10.0.32.71' หรือ '10.0.32.0/24'
    description     VARCHAR(200),
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by      VARCHAR(100),
    CONSTRAINT uq_sso_ip_env UNIQUE (env_id, ip_cidr)
);

-- 4) Allowed AD Users : รายชื่อ AD ที่อนุญาตให้เข้า env นั้น
--    ผูกกับ env_id (ไม่ผูกกับ IP) เพื่อให้รายชื่อ AD ใช้ร่วมได้หลาย env
CREATE TABLE IF NOT EXISTS sso_allowed_users (
    allowed_user_id SERIAL PRIMARY KEY,
    env_id          INT          NOT NULL REFERENCES sso_environments(env_id) ON DELETE CASCADE,
    ad_username     VARCHAR(100) NOT NULL,                -- 'somchai.s'
    employee_id     VARCHAR(50)  NULL,                    -- 'EMP001'
    display_name    VARCHAR(200) NULL,                    -- sync จาก AD
    email           VARCHAR(200) NULL,
    department      VARCHAR(100) NULL,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    granted_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    granted_by      VARCHAR(100) NULL,
    expires_at      TIMESTAMPTZ  NULL,                    -- กำหนดวันหมดอายุได้
    last_sync_at    TIMESTAMPTZ  NULL,                    -- sync AD ล่าสุด
    CONSTRAINT uq_sso_user_env UNIQUE (env_id, ad_username)
);

-- 5) Access Policies (ทางเลือก) : รวมเงื่อนไข IP + AD เป็น policy เดียว
--    ถ้าอยากให้ "user X เข้าได้เฉพาะจาก IP Y" ให้ใช้ตารางนี้
CREATE TABLE IF NOT EXISTS sso_access_policies (
    policy_id       SERIAL PRIMARY KEY,
    env_id          INT          NOT NULL REFERENCES sso_environments(env_id) ON DELETE CASCADE,
    policy_name     VARCHAR(100) NOT NULL,
    ip_cidr         INET         NULL,                    -- ถ้า NULL = ทุก IP ที่อยู่ใน sso_allowed_ips
    ad_username     VARCHAR(100) NULL,                    -- ถ้า NULL = ทุก user ที่อยู่ใน sso_allowed_users
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- 6) Login Audit Log : บันทึกทุกการขอเข้าระบบ (allowed / denied)
CREATE TABLE IF NOT EXISTS sso_login_audit (
    audit_id        BIGSERIAL PRIMARY KEY,
    app_code        VARCHAR(50),
    env_code        VARCHAR(20),
    base_url        VARCHAR(255),
    client_ip       INET,
    ad_username     VARCHAR(100),
    employee_id     VARCHAR(50),
    display_name    VARCHAR(200),
    result          VARCHAR(20)  NOT NULL,                -- 'ALLOW' | 'DENY_IP' | 'DENY_USER' | 'DENY_APP' | 'ERROR'
    deny_reason     VARCHAR(200),
    user_agent      TEXT,
    request_id      VARCHAR(64),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- 7) Active Sessions : session ที่ login สำเร็จ (สำหรับ middleware ตรวจซ้ำ)
CREATE TABLE IF NOT EXISTS sso_sessions (
    session_id      BIGSERIAL PRIMARY KEY,
    session_hash    CHAR(64)     NOT NULL UNIQUE,
    app_id          INT          NOT NULL REFERENCES sso_applications(app_id) ON DELETE CASCADE,
    env_id          INT          NOT NULL REFERENCES sso_environments(env_id) ON DELETE CASCADE,
    ad_username     VARCHAR(100) NOT NULL,
    employee_id     VARCHAR(50),
    client_ip       INET,
    user_agent      TEXT,
    status          VARCHAR(20)  NOT NULL DEFAULT 'ACTIVE', -- 'ACTIVE' | 'LOGOUT' | 'EXPIRED' | 'REVOKED'
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ  NOT NULL,
    logged_out_at   TIMESTAMPTZ  NULL
);

-- ============================================================
-- Indexes
-- ============================================================
CREATE INDEX IF NOT EXISTS ix_sso_env_app_active   ON sso_environments(app_id, is_active);
CREATE INDEX IF NOT EXISTS ix_sso_ip_env_active    ON sso_allowed_ips(env_id, is_active);
CREATE INDEX IF NOT EXISTS ix_sso_user_env_active  ON sso_allowed_users(env_id, is_active);
CREATE INDEX IF NOT EXISTS ix_sso_user_ad          ON sso_allowed_users(ad_username);
CREATE INDEX IF NOT EXISTS ix_sso_policy_env       ON sso_access_policies(env_id, is_active);
CREATE INDEX IF NOT EXISTS ix_sso_audit_created    ON sso_login_audit(created_at DESC);
CREATE INDEX IF NOT EXISTS ix_sso_audit_result     ON sso_login_audit(result, created_at DESC);
CREATE INDEX IF NOT EXISTS ix_sso_audit_user_time  ON sso_login_audit(ad_username, created_at DESC);
CREATE INDEX IF NOT EXISTS ix_sso_sessions_user    ON sso_sessions(ad_username, status, expires_at);
CREATE INDEX IF NOT EXISTS ix_sso_sessions_exp     ON sso_sessions(expires_at) WHERE status = 'ACTIVE';

-- ============================================================
-- Seed : HelpDesk (prod + test) ตามตัวอย่างใน requirement
-- ============================================================
INSERT INTO sso_applications (app_code, app_name, description)
VALUES ('HELPDESK', 'HelpDesk System', 'ระบบแจ้งซ่อม / HelpDesk')
ON CONFLICT (app_code) DO NOTHING;

INSERT INTO sso_environments (app_id, env_code, env_name, base_url, host_ip, base_path)
SELECT app_id, 'PROD', 'Production', 'http://10.0.32.71/HelpDesk/', '10.0.32.71'::inet, '/HelpDesk/'
FROM sso_applications WHERE app_code = 'HELPDESK'
ON CONFLICT (app_id, env_code) DO NOTHING;

INSERT INTO sso_environments (app_id, env_code, env_name, base_url, host_ip, base_path)
SELECT app_id, 'TEST', 'Test', 'http://10.115.2.61/HelpDesk/', '10.115.2.61'::inet, '/HelpDesk/'
FROM sso_applications WHERE app_code = 'HELPDESK'
ON CONFLICT (app_id, env_code) DO NOTHING;
