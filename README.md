# SSO Login - Permission Controller

## Overview

SSO Login เป็นบริการควบคุมสิทธิ์การเข้าถึง web application หลายระบบ โดยตรวจสิทธิ์จาก
**base URL + client IP + AD username** ผ่าน PostgreSQL ใช้เป็นเกตเวย์ด้านหน้า (หรือ middleware)
ก่อนปล่อยให้ผู้ใช้เข้าใช้งานระบบปลายทาง เช่น `HelpDesk`, `ERP`, `CRM` เป็นต้น

ปัญหาที่แก้:

* ผู้ดูแลต้องการจำกัดว่า "ผู้ใช้ AD คนไหน" เข้า "ระบบไหน" "ผ่าน IP/เครือข่ายไหน" ได้บ้าง
* ต้องแยก environment เช่น `PROD` (`http://10.0.32.71/HelpDesk/`) กับ `TEST` (`http://10.115.2.61/HelpDesk/`) ไม่ให้ปนกัน
* ต้องเพิ่ม/ลดรายชื่อผู้ใช้และ IP ได้จากฐานข้อมูล โดยไม่ต้อง redeploy
* ต้องมี audit log ทุกครั้งที่มีการขอเข้าระบบ (ALLOW/DENY)

## Objectives

* ให้บริการ REST API ตรวจสิทธิ์แบบ real-time ตาม (base URL, client IP, AD username)
* รองรับการลงทะเบียนหลาย application/environment
* จัดการรายชื่อ IP/CIDR และ AD users ผ่าน REST API
* บันทึก audit log ทุกครั้งที่ตรวจสิทธิ์
* พัฒนาด้วย Go ให้ทำงานเร็วและ deploy ง่าย

## Features

* ตรวจสิทธิ์ตาม **base URL + IP/CIDR + AD username**
* CRUD applications / environments / allowed IPs / allowed users
* รองรับทั้ง IP เดี่ยว (`10.0.32.71`) และ CIDR (`10.0.32.0/24`)
* กำหนดวันหมดอายุ (`expires_at`) ของสิทธิ์ผู้ใช้ได้
* Audit log ทุก request
* CORS + graceful shutdown
* ตั้งค่าผ่าน environment variables (`.env` ได้)

## Technology Stack

* Backend: Go 1.23+ (standard library `net/http` + `github.com/jackc/pgx/v5`)
* Database: PostgreSQL 13+ (ใช้ `INET` สำหรับเก็บ IP/CIDR)
* API: REST/JSON
* Migration: SQL ไฟล์ใน `migrations/`
* Gateway: โปรเจกต์ย่อย `api_gatewayGo/` (Go) สำหรับ reverse proxy

## Architecture

```text
[Client Browser]
    ↓  (http://10.0.32.71/HelpDesk/)
[Reverse Proxy / Gateway :18000]
    ↓
[SSO Login backend :12080]  ←── PostgreSQL (sso_permission)
    ↓  (ALLOW / DENY_*)
[Upstream App : HelpDesk / ERP / CRM ...]
```

Flow การตรวจสิทธิ์:

1. Gateway รับ request → ดึง `baseUrl` ของปลายทาง
2. Gateway เรียก `POST /api/check-access` ของ SSO Login พร้อม `baseUrl, clientIp, adUsername`
3. SSO Login ตรวจตามลำดับ: env → IP → user
4. ถ้า `ALLOW` gateway ส่งต่อไปยัง upstream, ถ้า `DENY_*` ตอบ 403
5. ทุก request ถูกบันทึกใน `sso_login_audit`

## Project Structure

```text
SSO-login/
├── README.md                    # เอกสารนี้
├── AGENTS.md                    # กฎสำหรับ AI agent
├── docs/
│   ├── project_rules.md         # กฎของโปรเจกต์
│   └── user_rules.md            # ความชอบของเจ้าของโปรเจกต์
├── migrations/
│   ├── 001_sso_permission_init.sql
│   └── 002_add_aduser_to_envs.sql
├── api_gatewayGo/               # Go reverse proxy (port 18000)
├── backend/                     # Go SSO permission service (port 12080)
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── model/
│   │   ├── store/               # PostgreSQL repository
│   │   ├── service/             # business logic
│   │   └── httpapi/             # HTTP handlers
│   ├── go.mod
│   ├── .env.example
│   ├── API.md
│   └── README.md
└── frontend/                    # React + Vite + TailwindCSS (port 3000)
    ├── src/
    │   ├── pages/               # Login, Dashboard
    │   ├── components/          # EnvManager, AppManager, ...
    │   └── api.js
    ├── package.json
    ├── tailwind.config.js
    ├── vite.config.js
    └── README.md
```

## Requirements

* Go 1.23+
* PostgreSQL 13+
* Windows / Linux
* ไม่ต้องใช้ dependency ภายนอกอื่นนอกจาก `pgx/v5`

## Installation

```bash
# clone & install backend
cd backend
go mod tidy
go build ./...
```

สร้าง database และ schema:

```bash
psql -U postgres -c "CREATE DATABASE sso_permission;"
psql -U postgres -d sso_permission -f ../migrations/001_sso_permission_init.sql
```

## Configuration

ใช้ไฟล์ `.env` ในโฟลเดอร์ `backend/` (ดูตัวอย่างใน [backend/.env.example](./backend/.env.example)):

```text
HTTP_ADDR=:12080
FRONTEND_ORIGIN=*
DATABASE_URL=postgres://sso:CHANGE_ME@127.0.0.1:5432/sso_permission?sslmode=disable
DB_MAX_OPEN=20
DB_MAX_LIFETIME_MIN=30
```

ห้าม commit secrets จริงลง Git

## Development

```bash
cd backend
go run ./cmd/server
```

Service จะฟังที่ `:12080` (override ได้ด้วย `HTTP_ADDR`)

## Testing

```bash
cd backend
go vet ./...
go build ./...
```

## Build

```bash
cd backend
go build -o bin/sso-login-server ./cmd/server
```

## Deployment

* Build เป็น Windows service หรือรันด้วย `nssm` / `sc.exe`
* ใช้ env file หรือ Windows Environment Variables
* เปิด port 12080 บน firewall เฉพาะ network ที่ต้องการ

## API

ดูรายละเอียดที่ [backend/API.md](./backend/API.md)

Endpoint หลัก:

* `POST /api/check-access` — ตรวจสิทธิ์
* CRUD `/api/apps`, `/api/envs`, `/api/allowed-ips`, `/api/allowed-users`
* `GET /api/audit` — ประวัติการขอเข้าระบบ

## Troubleshooting

* **`missing required environment variables: DATABASE_URL`**
  → ตั้ง `DATABASE_URL` ใน `.env` หรือ environment
* **`db open: ... connection refused`**
  → ตรวจ PostgreSQL เปิดอยู่และ `DATABASE_URL` ถูกต้อง
* **`DENY_APP` ทุก request**
  → ตรวจว่า `baseUrl` ใน `sso_environments` ตรงกับที่ client เรียกเป๊ะ (รวม `http://` และ `/` ท้าย)

## Contributing

Pull request ยินดี แต่ขอให้ทำตามกฎใน [docs/project_rules.md](./docs/project_rules.md)

## License

ภายในองค์กร (internal use)

## Documentation

* [AGENTS.md](./AGENTS.md) — กฎสำหรับ AI agent
* [docs/project_rules.md](./docs/project_rules.md) — กฎของโปรเจกต์
* [docs/user_rules.md](./docs/user_rules.md) — ความชอบของเจ้าของโปรเจกต์
* [backend/API.md](./backend/API.md) — API reference
* [backend/README.md](./backend/README.md) — รายละเอียด backend
* [migrations/](./migrations/) — SQL migrations
