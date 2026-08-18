# SSO Login - Docker Deployment Guide

คู่มือการ deploy ระบบ SSO Login ด้วย Docker

---

## สถาปัตยกรรม

```
┌─────────────────────────────────────────────────────────────┐
│                    Docker Network                            │
│                                                             │
│  ┌──────────────┐    ──────────────┐    ┌──────────────┐  │
│  │   Frontend   │───▶│    Gateway   │───▶│   Backend    │  │
│  │   (Nginx)    │    │   (Go)       │    │   (Go)       │  │
│  │   Port :80   │    │  Port :18000 │    │  Port :12080 │  │
│  └──────────────┘    ──────────────┘    └─────────────┘  │
│                                                  │          │
│                                          ┌───────▼───────┐  │
│                                          │   PostgreSQL   │  │
│                                          │   Port :5432   │  │
│                                          └───────────────  │
└─────────────────────────────────────────────────────────────┘
```

---

## ข้อกำหนดเบื้องต้น

- Docker Desktop ติดตั้งแล้ว
- Docker Compose v2 ขึ้นไป

ตรวจสอบเวอร์ชัน:
```bash
docker --version
docker compose version
```

---

## วิธีรันระบบ

### 1. Build และรันทั้งหมด

```bash
docker compose up -d --build
```

### 2. ดูสถานะ services

```bash
docker compose ps
```

### 3. ดู logs

```bash
# ดู logs ทั้งหมด
docker compose logs -f

# ดู logs ของ service เฉพาะ
docker compose logs -f backend
docker compose logs -f api_gateway
docker compose logs -f frontend
```

### 4. หยุดระบบ

```bash
docker compose down
```

### 5. หยุดและลบ volumes (reset database)

```bash
docker compose down -v
```

---

## การเข้าถึงระบบ

| Service | URL | Port |
|---------|-----|------|
| Frontend | http://localhost | 80 |
| API Gateway | http://localhost:18000 | 18000 |
| Backend API | http://localhost:12080 | 12080 |
| PostgreSQL | localhost | 5432 |

---

## การตั้งค่า Environment Variables

### Backend

แก้ไขใน `docker-compose.yml` ส่วน `backend.environment`:

```yaml
environment:
  - HTTP_ADDR=0.0.0.0:12080
  - FRONTEND_ORIGIN=*
  - DATABASE_URL=postgres://pguser:pgpass123@postgres:5432/postgres?sslmode=disable
  - DB_MAX_OPEN=20
  - DB_MAX_LIFETIME_MIN=30
  - JWT_SECRET=<your-secret-here>
  - JWT_TTL_MINUTES=480
```

### API Gateway

แก้ไขใน `docker-compose.yml` ส่วน `api_gateway.environment`:

```yaml
environment:
  - GATEWAY_ADDR=0.0.0.0:18000
  - BACKEND_URL=http://backend:12080
```

### Frontend

Frontend ใช้ relative path `/api/SSO_login` โดยอัตโนมัติ
Nginx จะ proxy ไปที่ gateway

ถ้าต้องการเปลี่ยน API base URL:
แก้ไข `frontend/nginx.conf` ส่วน `location /api/SSO_login/`

---

## Health Check

ระบบมี health check อัตโนมัติสำหรับทุก service:

```bash
# ตรวจสอบ health ของแต่ละ service
docker compose ps

# หรือเรียก health endpoint โดยตรง
curl http://localhost:12080/health
curl http://localhost:18000/health
curl http://localhost:80
```

---

## การจัดการ Database

### เข้าถึง PostgreSQL โดยตรง

```bash
docker compose exec postgres psql -U pguser -d postgres
```

### Backup Database

```bash
docker compose exec postgres pg_dump -U pguser postgres > backup.sql
```

### Restore Database

```bash
docker compose exec -T postgres psql -U pguser postgres < backup.sql
```

### รัน Migrations

Migrations จะถูกรันอัตโนมัติตอน first start จากโฟลเดอร์ `migrations/`

ถ้าต้องการรัน migrations ใหม่:
```bash
docker compose down -v
docker compose up -d
```

---

## Production Checklist

- [ ] เปลี่ยน `JWT_SECRET` เป็นค่า random 32+ bytes
- [ ] เปลี่ยน `POSTGRES_PASSWORD` เป็นรหัสผ่านที่แข็งแรง
- [ ] ตั้งค่า `FRONTEND_ORIGIN` ให้เฉพาะเจาะจง (ไม่ใช่ `*`)
- [ ] ใช้ HTTPS (เพิ่ม reverse proxy เช่น Traefik หรือ Nginx ด้านหน้า)
- [ ] ตั้งค่า resource limits ใน docker-compose.yml
- [ ] ตั้งค่า logging driver (เช่น json-file, syslog)
- [ ] Backup database เป็นประจำ

### ตัวอย่าง Production docker-compose.yml

```yaml
services:
  backend:
    environment:
      - JWT_SECRET=a1b2c3d4e5f6...  # เปลี่ยนเป็นค่าจริง
      - FRONTEND_ORIGIN=https://sso.example.com
  
  postgres:
    environment:
      POSTGRES_PASSWORD: <strong-password-here>  # เปลี่ยนเป็นรหัสผ่านที่แข็งแรง
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '1.0'
```

---

## Troubleshooting

### ปัญหา: Container ไม่ start

```bash
# ดู logs
docker compose logs backend

# ตรวจสอบ health
docker compose ps

# ลอง build ใหม่
docker compose up -d --build
```

### ปัญหา: Database connection error

```bash
# ตรวจสอบว่า postgres รันอยู่
docker compose ps postgres

# ดู logs ของ postgres
docker compose logs postgres

# ลองเชื่อมต่อจาก backend
docker compose exec backend wget -qO- http://postgres:5432
```

### ปัญหา: Port already in use

```bash
# ตรวจสอบ port ที่ใช้งานอยู่
netstat -an | findstr "80 12080 18000 5432"

# เปลี่ยน port ใน docker-compose.yml
ports:
  - "8080:80"  # เปลี่ยนจาก 80 เป็น 8080
```

### ปัญหา: Frontend ไม่สามารถเรียก API ได้

ตรวจสอบ `frontend/nginx.conf`:
```nginx
location /api/SSO_login/ {
    proxy_pass http://api_gateway:18000/api/SSO_login/;
    # ตรวจสอบว่า api_gateway ตรงกับ service name ใน docker-compose.yml
}
```

---

## การ Update ระบบ

```bash
# Pull code ใหม่
git pull

# Build และ restart
docker compose up -d --build

# ดู logs
docker compose logs -f
```

---

## การลบระบบทั้งหมด

```bash
# หยุดและลบ containers, networks
docker compose down

# ลบ volumes (database จะถูกลบ)
docker compose down -v

# ลบ images
docker compose down --rmi all
```

---

## ไฟล์ที่เกี่ยวข้อง

```
.
├── docker-compose.yml          # Orchestration
── .dockerignore               # Docker ignore rules
├── backend/
│   ├── Dockerfile              # Backend image
│   └── .env.example            # Backend env template
├── api_gatewayGo/
│   ├── Dockerfile              # Gateway image
│   └── .env.example            # Gateway env template
├── frontend/
│   ├── Dockerfile              # Frontend image
│   └── nginx.conf              # Nginx configuration
└── migrations/                 # Database migrations
```

---

## คำสั่ง Docker ที่ใช้บ่อย

```bash
# Build images
docker compose build

# Start services
docker compose up -d

# Stop services
docker compose down

# View logs
docker compose logs -f [service]

# Execute command in container
docker compose exec backend sh
docker compose exec postgres psql -U pguser

# View running containers
docker compose ps

# Restart specific service
docker compose restart backend

# View resource usage
docker stats
```

---

**Version:** 1.0  
**Last Updated:** 2026-08-18
