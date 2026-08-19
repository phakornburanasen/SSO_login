# Project Rules

## Purpose

This document defines project-specific rules, conventions, and technical constraints
for the **SSO Login — Permission Controller** project.

These rules apply to all contributors and AI coding agents working on this project.

---

## 1. General Rules

* Follow the existing project architecture (see [README.md § Architecture](../README.md#architecture)).
* Prefer simple and maintainable solutions.
* Reuse existing functionality where appropriate.
* Do not create duplicate functionality.
* Keep changes focused and minimal.
* Do not modify unrelated code.
* Preserve existing behavior unless a change is explicitly required.
* สื่อสารกับผู้ใช้เป็นภาษาไทย ยกเว้น technical terms ที่นิยมเขียนเป็นภาษาอังกฤษ

---

## 2. Architecture

The project has two Go services + a PostgreSQL database:

```text
[Client]
    ↓
[Reverse Proxy / Gateway :18000]  (api_gatewayGo/)
    ↓
[SSO Login backend :12080]        (backend/)
    ↓
[PostgreSQL — sso_permission]
```

Rules:

* **backend** เป็นเจ้าของ business logic การตรวจสิทธิ์ (check-access, CRUD) — ห้ามย้ายไปอยู่ gateway
* **gateway** ทำหน้าที่ reverse proxy เท่านั้น — ห้ามฝัง business logic
* **database** เป็นแหล่งเดียวของ IP whitelist, AD users และ audit log
* Do not introduce a new architectural pattern without a clear reason.
* Keep responsibilities separated between components.

---

## 3. Project Structure

```text
backend/                     Go service (port 12080)
    cmd/server/main.go
    internal/
        config/              โหลด env + .env
        model/               DTO/struct
        store/               PostgreSQL repository (pgx/v5)
        service/             business logic (check-access)
        httpapi/             HTTP handlers + middleware
    .env.example
    go.mod

api_gatewayGo/               Go reverse proxy (port 18000)
migrations/                  SQL migrations (PostgreSQL)
docs/                        project + user rules
```

Keep this section updated when the structure changes significantly.

---

## 4. Coding Standards

* ใช้ Go standard style (`gofmt`)
* ใช้ `net/http` ของ standard library — ไม่เพิ่ม web framework ถ้าไม่จำเป็น
* แยก concerns: `model` (data) → `store` (DB) → `service` (logic) → `httpapi` (transport)
* ใช้ `context.Context` ตลอดทุก call ที่ติดต่อ DB
* Validate input ใน `service` layer ไม่ใช่แค่ handler
* ห้ามใช้ `panic` ใน business code — ใช้ `error` return
* Add comments only when they explain non-obvious behavior.
* Do not add comments that simply repeat the code.

---

## 5. Naming Conventions

* Go: ใช้ `camelCase` สำหรับ variable/function, `PascalCase` สำหรับ exported, `UPPER_CASE` สำหรับ const
* Database: snake_case (เช่น `sso_login_audit`, `host_ip`)

### Login permission model (สำคัญ)

แยก **admin role** ออกจาก **env ownership** เสมอ — ห้ามปนกัน:

| Role | ตรวจจาก | สิทธิ์ |
|------|---------|--------|
| **admin** | `sso_admins` (ตารางแยก) | เห็น env ทั้งหมด, จัดการ admins |
| **user** | `sso_environments.ADUser` | เฉพาะ env ที่ตัวเองเป็น ADUser |

ลำดับการตรวจสอบ login:
1. username/password ไม่ว่าง
2. `IsAdminUser` → query `sso_admins` เท่านั้น (ห้าม query `sso_environments`)
3. ถ้าไม่ใช่ admin → `ListAccessibleEnvsByADUser`
4. ถ้า envs ว่าง → `ErrNoPermission` (HTTP 403)
5. ถ้า pass ทั้งหมด → ออก JWT ใส่ role ใน claims

* JSON tag: camelCase (เช่น `adUsername`, `baseUrl`)
* API path: kebab-case (เช่น `/api/check-access`, `/api/allowed-users`)

---

## 6. API Rules

* ใช้ HTTP method ตาม REST convention
* Return JSON เท่านั้น — ไม่ส่ง HTML / text
* Error response ต้องมี `{"error": "..."}` เสมอ
* Validation error ใช้ `400`, not found ใช้ `404`, conflict ใช้ `409`, forbidden ใช้ `403`
* ห้ามสร้าง endpoint ที่ซ้ำซ้อน
* ทุก endpoint ที่แก้ไข state ต้องมี audit log (ผ่าน service layer)

ดูรายละเอียดได้ที่ [backend/API.md](../backend/API.md)

---

## 7. Database Rules

* ใช้ PostgreSQL 13+
* ใช้ `INET` สำหรับเก็บ IP/CIDR — ห้ามเก็บเป็น string
* ใช้ `TIMESTAMPTZ` เสมอ — ห้ามใช้ `TIMESTAMP` (without tz)
* ใช้ migration files ใน `migrations/` — ห้ามแก้ schema ผ่าน psql โดยตรง
* ห้าม hard-code credentials ใน source code
* ห้าม drop / truncate ตาราง production โดยไม่ได้รับอนุมัติ
* Verify the scope of `UPDATE` and `DELETE` operations before execution.

---

## 8. Security Rules

* ห้าม hard-code password / API key / token ใน source code
* ห้าม commit secrets ลง Git
* Validate และ sanitize external input ทุก field
* ใช้ prepared statement เสมอ (pgx ทำให้อัตโนมัติ)
* ห้าม log password, token หรือ sensitive info
* IP/CIDR ต้องผ่าน `netip.ParsePrefix` หรือ `netip.ParseAddr` ก่อนใช้
* baseUrl ต้องตรงกันแบบ exact match เพื่อป้องกัน path confusion

---

## 9. Dependencies

* ใช้ dependency น้อยที่สุด — ปัจจุบันมีแค่ `github.com/jackc/pgx/v5`
* ห้ามเพิ่ม dependency ใหม่ถ้า standard library ทำได้
* ถ้าจำเป็นต้องเพิ่ม ต้องอธิบายเหตุผลก่อน
* ใช้ version ที่ compatible กับ Go 1.23+

---

## 10. Configuration

* ใช้ environment variables ทั้งหมด — ไม่ hard-code ค่าใน source
* ใช้ไฟล์ `.env` สำหรับ local development (โหลดผ่าน `config.Load()`)
* ห้าม commit `.env` ที่มี secrets จริง — ใช้ `.env.example` เป็น template
* Document required environment variables ใน [backend/.env.example](../backend/.env.example)

Environment variables ที่ใช้:

| Variable             | Required | คำอธิบาย                       |
|----------------------|----------|---------------------------------|
| `HTTP_ADDR`          | no       | default `:12080`                |
| `FRONTEND_ORIGIN`    | no       | default `*` (CORS)              |
| `DATABASE_URL`       | **yes**  | pgx connection string           |
| `DB_MAX_OPEN`        | no       | default `20`                    |
| `DB_MAX_LIFETIME_MIN`| no       | default `30`                    |

---

## 11. Testing

* รัน `go vet ./...` และ `go build ./...` ก่อน commit
* ถ้ามี unit test ให้รัน `go test ./...`
* ห้ามลบ test เพื่อให้ test ผ่าน — แก้ที่ root cause
* ถ้า test รันไม่ได้ให้อธิบายเหตุผล

---

## 12. Error Handling and Logging

* Handle expected errors explicitly — ห้าม `log.Fatal` ใน handler
* ใช้ `log.Printf` ของ standard library
* Log request ใน middleware: method, path, ip, ua, duration
* ห้าม log password, token, secrets
* Return error message ที่ปลอดภัยต่อผู้ใช้ — ไม่เปิดเผย internal stack

---

## 13. Performance

* ใช้ connection pool (`pgxpool`) แล้ว — ห้ามสร้าง connection ใหม่ต่อ request
* `MAX_CONNS = 20`, `MAX_CONN_LIFETIME = 30 นาที` (override ได้)
* ใช้ index ที่กำหนดใน migration เพื่อให้ query เร็ว
* หลีกเลี่ยง `SELECT *` — ระบุ field ที่ต้องการ

---

## 14. Git

* ใช้ commit message ที่สื่อความหมาย เช่น `feat: add check-access endpoint`
* Keep commits focused
* ห้าม force push โดยไม่ได้รับอนุมัติ
* ห้าม discard unrelated user changes

---

## 15. Deployment

* Environment: Production / Staging / Development
* Deployment method: Windows service ผ่าน `nssm` หรือ binary ตรง
* Required services: PostgreSQL 13+
* Required env vars: ดูหัวข้อ Configuration

---

## 16. Project-Specific Rules

* `baseUrl` ใน request ต้องตรงกับ `sso_environments.base_url` แบบ exact match
* การตรวจสิทธิ์ทำตามลำดับ: `env → IP → user` ถ้า fail ที่ขั้นไหนให้หยุดทันที
* `result` ของ audit log ต้องเป็น enum: `ALLOW | DENY_APP | DENY_IP | DENY_USER | ERROR`
* ห้ามเปลี่ยน schema โดยไม่เพิ่ม migration ใหม่
* ใช้ `INET` ของ PostgreSQL เพื่อเช็ค `<<=` (ip contained in cidr)

---

## 17. Change Policy

When changing an established rule:

1. Update this document.
2. Update affected documentation.
3. Verify affected code.
4. Inform contributors when necessary.

The documentation should reflect the actual project behavior.
