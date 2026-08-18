# sso-login backend (Go)

บริการตรวจสิทธิ์การเข้าระบบตาม **base URL + client IP + AD username** สำหรับหลายแอปพลิเคชัน
ออกแบบให้เก็บกฎ IP/ผู้ใช้ AD ลงฐานข้อมูล PostgreSQL แล้วให้ middleware ภายนอกเรียก `POST /api/check-access`

## Port

* HTTP: **`:12010`** (override ได้ด้วย `HTTP_ADDR` env)

## โครงสร้าง

```text
backend/
├── cmd/server/main.go
├── internal/
│   ├── config/config.go     # โหลด .env + env vars
│   ├── model/model.go       # DTO/struct
│   ├── store/store.go       # PostgreSQL (pgx/v5) repository
│   ├── service/service.go   # business logic (check-access)
│   └── httpapi/api.go       # HTTP handlers + middleware
├── go.mod
├── .env.example
├── API.md
└── README.md
```

## ติดตั้ง

```bash
cd backend
go mod tidy
go build ./...
```

## รัน

```bash
# 1) ใช้ schema ที่ ../migrations/001_sso_permission_init.sql
psql -h 127.0.0.1 -U sso -d sso_permission -f ../migrations/001_sso_permission_init.sql

# 2) ตั้ง env (หรือใช้ .env)
set HTTP_ADDR=:12010
set DATABASE_URL=postgres://sso:sso_password@127.0.0.1:5432/sso_permission?sslmode=disable

# 3) รัน
go run ./cmd/server
```

## API สรุป

| Method | Path                    | ใช้ทำอะไร                                  |
|--------|-------------------------|--------------------------------------------|
| GET    | `/health`               | health check                               |
| POST   | `/api/check-access`     | ตรวจสิทธิ์ `{baseUrl,clientIp,adUsername}`  |
| GET    | `/api/apps`             | list applications                          |
| POST   | `/api/apps`             | create application                         |
| PUT    | `/api/apps/{id}`        | update application                         |
| DELETE | `/api/apps/{id}`        | delete application                         |
| GET    | `/api/envs?appId=`      | list environments                          |
| POST   | `/api/envs`             | create environment                         |
| PUT    | `/api/envs/{id}`        | update environment                         |
| DELETE | `/api/envs/{id}`        | delete environment                         |
| GET    | `/api/allowed-ips?envId=` | list allowed IPs/CIDR                    |
| POST   | `/api/allowed-ips`      | add IP/CIDR (เช่น `10.0.32.0/24`)           |
| DELETE | `/api/allowed-ips/{id}` | remove IP                                  |
| GET    | `/api/allowed-users?envId=` | list allowed AD users                   |
| POST   | `/api/allowed-users`    | add AD user                                |
| DELETE | `/api/allowed-users/{id}` | remove AD user                          |
| GET    | `/api/audit?limit=100`  | ดูประวัติการขอเข้าระบบ                      |

รายละเอียด payload/response ดูได้ที่ [API.md](./API.md)
