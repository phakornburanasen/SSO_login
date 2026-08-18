# Frontend (SSO_login) — Vite + React + TailwindCSS

UI สำหรับจัดการ SSO Permission Controller

* **Port**: `3000` (Vite default)
* **URL base**: `/SSO_login/` (assets, JS, CSS โหลดจาก path นี้)
* **API base**: `http://127.0.0.1:18000/api/SSO_login` (เรียกตรงไปที่ `api_gatewayGo` port 18000)
* **Stack**: Vite 5, React 18, TailwindCSS 3

## Features

* Login page → เก็บ token ใน `localStorage`
* Dashboard พร้อม 6 เมนู:
  * **สภาพแวดล้อม** — CRUD `sso_environments` (env_name, base_url, host_ip, base_path, is_active, ADUser)
  * **แอปพลิเคชัน** — CRUD `sso_applications`
  * **IP ที่อนุญาต** — CRUD `sso_allowed_ips`
  * **ผู้ใช้ AD** — CRUD `sso_allowed_users`
  * **ทดสอบสิทธิ์** — เรียก `POST /api/check-access` ตรงๆ
  * **Audit Log** — ดู `sso_login_audit`

## Install & Run

```bash
cd frontend
npm install
npm run dev          # dev server :3000 — เข้า http://localhost:3000/SSO_login/
npm run build        # build production → dist/
npm run preview      # serve dist on :3000
```

## Environment

สร้างไฟล์ `.env` (ไม่บังคับ — มี default ให้แล้ว):

```text
# API base path ที่ frontend ใช้เรียกตรงไปที่ api_gatewayGo (port 18000)
# - Dev  : http://127.0.0.1:18000/api/SSO_login
# - Prod : http://<gateway-host>:18000/api/SSO_login
VITE_API_BASE=http://127.0.0.1:18000/api/SSO_login
```

## URL Convention

ทั้งหมดต้องตรงกัน:

| ที่ไหน                       | ค่า                  | หมายเหตุ |
|------------------------------|----------------------|---------|
| `api_gatewayGo/service.json` | `"SSO_login": "http://127.0.0.1:12010"` | service name → backend |
| Gateway port                 | `18000`              | หลีกเลี่ยง ERR_UNSAFE_PORT ของ Chrome/Edge |
| Vite `base`                  | `/SSO_login/`        | path ที่ assets โหลด |
| `VITE_API_BASE`              | `http://127.0.0.1:18000/api/SSO_login` | API base |
| เข้าใช้งาน (prod)            | `http://gateway:18000/SSO_login/` | ผ่าน gateway |
| เข้าใช้งาน (dev)             | `http://localhost:3000/SSO_login/` | Vite dev server |

## Flow การเชื่อมต่อ

**Production** (เสิร์ฟผ่าน gateway):

```text
Browser  →  http://gateway:18000/SSO_login/         (ไฟล์ static)
Browser  →  http://gateway:18000/api/SSO_login/api/auth/login
        ↓
[Gateway :18000]
        ↓  rewrite → http://127.0.0.1:12010/api/auth/login
[SSO backend :12010]
        ↓
[PostgreSQL @ 10.0.32.71]
```

**Development** (`npm run dev`):

```text
Browser  →  http://localhost:3000/SSO_login/        (Vite serves)
Browser  →  http://127.0.0.1:18000/api/SSO_login/api/auth/login
        ↓
[Gateway :18000]
        ↓  rewrite → http://127.0.0.1:12010/api/auth/login
[SSO backend :12010]
```

## โครงสร้าง

```text
frontend/
├── index.html
├── package.json
├── vite.config.js
├── tailwind.config.js
├── postcss.config.js
└── src/
    ├── main.jsx
    ├── App.jsx
    ├── index.css         # Tailwind base + components
    ├── api.js            # fetch wrapper + token
    ├── pages/
    │   ├── Login.jsx
    │   └── Dashboard.jsx
    └── components/
        ├── EnvManager.jsx        # CRUD sso_environments
        ├── AppManager.jsx        # CRUD sso_applications
        ├── AllowedIpManager.jsx  # CRUD sso_allowed_ips
        ├── AllowedUserManager.jsx# CRUD sso_allowed_users
        ├── AuditLog.jsx
        └── CheckAccess.jsx
```
