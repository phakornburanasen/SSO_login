# SSO Login - Deployment Guide

คู่มือการ deploy ระบบ SSO Login สำหรับสภาพแวดล้อมต่างๆ

---

## สถาปัตยกรรมระบบ

```
┌─────────────────────────────────────────────────────────────┐
│                      Client Browser                          │
│                  (แอปพลิเคชันที่ต้องการใช้ SSO)               │
└────────────────────────────┬────────────────────────────────┘
                             │ HTTP/HTTPS
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                  Web Server (Nginx/Apache)                   │
│  - Serve Frontend Static Files                               │
│  - Reverse Proxy: /api/SSO_login -> Gateway                  │
────────────────────────────┬────────────────────────────────
                             │ HTTP
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                  API Gateway (Port 18000)                    │
│  - Reverse Proxy                                             │
│  - Strip prefix: /api/SSO_login                              │
│  - Forward to Backend                                        │
└────────────────────────────┬────────────────────────────────┘
                             │ HTTP
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                  Backend API (Port 12080)                    │
│  - JWT Authentication                                        │
│  - Access Control Logic                                      │
│  - PostgreSQL Database                                       │
└─────────────────────────────────────────────────────────────┘
```

---

## Scenario 1: Single Server (แนะนำสำหรับ Production)

**เหมาะสำหรับ:** ระบบที่ frontend, gateway, backend อยู่บน server เดียวกัน

### ข้อดี
- ✅ ไม่มีปัญหา CORS (same-origin)
- ✅ ไม่ต้องตั้งค่า `VITE_API_BASE`
- ✅ ง่ายต่อการจัดการ
- ✅ ปลอดภัยกว่า (ไม่ expose port 12080, 18000 ออกภายนอก)

### การตั้งค่า

#### 1. Backend Configuration

```bash
cd backend
cp .env.example .env
```

แก้ไข `backend/.env`:
```env
HTTP_ADDR=127.0.0.1:12080
FRONTEND_ORIGIN=*
DATABASE_URL=postgres://pguser:pgpass123@localhost:5432/postgres?sslmode=disable
DB_MAX_OPEN=20
DB_MAX_LIFETIME_MIN=30
JWT_SECRET=<generate-random-32-bytes>
JWT_TTL_MINUTES=480
```

#### 2. API Gateway Configuration

```bash
cd api_gatewayGo
cp .env.example .env
```

แก้ไข `api_gatewayGo/.env`:
```env
GATEWAY_ADDR=127.0.0.1:18000
BACKEND_URL=http://127.0.0.1:12080
```

#### 3. Frontend Build

```bash
cd frontend

# สร้าง .env (ไม่จำเป็นต้องตั้ง VITE_API_BASE เพราะจะใช้ relative path)
cp .env.example .env

# Build สำหรับ production
npm run build
```

ผลลัพธ์จะอยู่ใน `frontend/dist/`

#### 4. Nginx Configuration

```nginx
server {
    listen 80;
    server_name sso.example.com;  # เปลี่ยนเป็น domain ของคุณ

    # Frontend static files
    location / {
        root /var/www/sso-frontend/dist;
        try_files $uri $uri/ /index.html;
        
        # Cache static assets
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }
    }

    # API Gateway proxy
    location /api/SSO_login/ {
        proxy_pass http://127.0.0.1:18000/api/SSO_login/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeout settings
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}
```

#### 5. systemd Services (Optional)

สร้าง service files สำหรับ auto-start:

**`/etc/systemd/system/sso-backend.service`:**
```ini
[Unit]
Description=SSO Login Backend
After=network.target postgresql.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/sso-login/backend
ExecStart=/opt/sso-login/backend/bin/sso-login-server
Restart=always
RestartSec=5
EnvironmentFile=/opt/sso-login/backend/.env

[Install]
WantedBy=multi-user.target
```

**`/etc/systemd/system/sso-gateway.service`:**
```ini
[Unit]
Description=SSO Login API Gateway
After=network.target sso-backend.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/sso-login/api_gatewayGo
ExecStart=/opt/sso-login/api_gatewayGo/bin/sso-login-gateway
Restart=always
RestartSec=5
EnvironmentFile=/opt/sso-login/api_gatewayGo/.env

[Install]
WantedBy=multi-user.target
```

Enable services:
```bash
sudo systemctl enable sso-backend
sudo systemctl enable sso-gateway
sudo systemctl start sso-backend
sudo systemctl start sso-gateway
```

---

## Scenario 2: Separate Servers (Frontend แยกจาก Backend)

**เหมาะสำหรับ:** Frontend อยู่บน CDN หรือ static hosting, Backend อยู่บน server อื่น

### การตั้งค่า

#### 1. Backend & Gateway (Server A: 10.0.32.71)

ตั้งค่าเหมือน Scenario 1 แต่เปลี่ยน:

`backend/.env`:
```env
HTTP_ADDR=0.0.0.0:12080
FRONTEND_ORIGIN=https://app.example.com  # domain ของ frontend
```

`api_gatewayGo/.env`:
```env
GATEWAY_ADDR=0.0.0.0:18000
BACKEND_URL=http://127.0.0.1:12080
```

#### 2. Frontend (Server B หรือ Static Hosting)

สร้าง `frontend/.env`:
```env
VITE_API_BASE=https://sso-api.example.com/api/SSO_login
```

Build:
```bash
npm run build
```

Deploy `frontend/dist/` ไปยัง static hosting (S3, CloudFront, Vercel, etc.)

#### 3. CORS Configuration

Backend จะส่ง CORS headers ตาม `FRONTEND_ORIGIN` ที่ตั้งค่าไว้

ตรวจสอบ response headers:
```bash
curl -I -X OPTIONS \
  -H "Origin: https://app.example.com" \
  -H "Access-Control-Request-Method: POST" \
  https://sso-api.example.com/api/SSO_login/api/check-access
```

Expected headers:
```
Access-Control-Allow-Origin: https://app.example.com
Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Allow-Credentials: true
```

---

## Scenario 3: Development Environment

**เหมาะสำหรับ:** นักพัฒนาที่ต้องการรันทุก service บนเครื่องตัวเอง

### วิธีที่ 1: ใช้ runserver.bat (Windows)

```bash
# รันทั้ง 3 service พร้อมกัน
.\runserver.bat
```

### วิธีที่ 2: รันแยกแต่ละ service

**Terminal 1 - Backend:**
```bash
cd backend
go run ./cmd/server
```

**Terminal 2 - Gateway:**
```bash
cd api_gatewayGo
go run ./cmd/server
```

**Terminal 3 - Frontend:**
```bash
cd frontend
npm run dev
```

### วิธีที่ 3: ใช้ Docker (ถ้ามี Dockerfile)

```bash
docker-compose up -d
```

---

## การตั้งค่า Firewall

### Windows Firewall

```powershell
# เปิด port สำหรับ Gateway (ถ้าต้องการให้เครื่องอื่นเข้าถึง)
New-NetFirewallRule -DisplayName "SSO Login Gateway" `
  -Direction Inbound `
  -LocalPort 18000 `
  -Protocol TCP `
  -Action Allow

# เปิด port สำหรับ Backend (ถ้าต้องการให้เข้าถึงตรง)
New-NetFirewallRule -DisplayName "SSO Login Backend" `
  -Direction Inbound `
  -LocalPort 12080 `
  -Protocol TCP `
  -Action Allow
```

### Linux (UFW)

```bash
sudo ufw allow 18000/tcp  # Gateway
sudo ufw allow 12080/tcp  # Backend (optional)
sudo ufw allow 80/tcp     # HTTP
sudo ufw allow 443/tcp    # HTTPS
```

---

## การตรวจสอบระบบ

### Health Check

```bash
# ตรวจสอบ Backend
curl http://localhost:12080/health

# ตรวจสอบ Gateway
curl http://localhost:18000/api/SSO_login/health
```

Expected response:
```json
{"status": "ok", "service": "sso-login"}
```

### ทดสอบ Login

```bash
curl -X POST http://localhost:18000/api/SSO_login/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'
```

### ทดสอบ Access Check

```bash
curl -X POST http://localhost:18000/api/SSO_login/api/check-access \
  -H "Content-Type: application/json" \
  -d '{
    "baseUrl": "http://10.0.32.71/HelpDesk/",
    "clientIp": "192.168.1.100",
    "adUsername": "somchai.s"
  }'
```

---

## Troubleshooting

### ปัญหา: CORS Error ใน Browser

**อาการ:**
```
Access to fetch at 'http://...' from origin 'http://...' 
has been blocked by CORS policy
```

**วิธีแก้:**
1. ตรวจสอบ `FRONTEND_ORIGIN` ใน `backend/.env`
2. ถ้า frontend และ backend อยู่คนละ domain ต้องตั้งค่าให้ตรงกัน
3. ตรวจสอบว่า Gateway ไม่ได้เพิ่ม CORS headers ซ้ำ (ดู logs)

### ปัญหา: Connection Refused

**อาการ:**
```
Unable to connect to API Gateway
```

**วิธีแก้:**
1. ตรวจสอบว่า service รันอยู่:
   ```bash
   # Windows
   netstat -an | findstr "18000 12080"
   
   # Linux
   ss -tlnp | grep -E "18000|12080"
   ```

2. ตรวจสอบ firewall rules
3. ตรวจสอบ `BACKEND_URL` ใน `api_gatewayGo/.env`

### ปัญหา: 401 Unauthorized

**อาการ:**
```json
{"error": "invalid or expired session"}
```

**วิธีแก้:**
1. ตรวจสอบว่า JWT token ถูกส่งใน header:
   ```
   Authorization: Bearer <token>
   ```

2. ตรวจสอบว่า token ยังไม่หมดอายุ (default 8 ชั่วโมง)
3. Login ใหม่เพื่อรับ token ใหม่

### ปัญหา: Database Connection Error

**อาการ:**
```
db open: failed to connect to PostgreSQL
```

**วิธีแก้:**
1. ตรวจสอบ `DATABASE_URL` ใน `backend/.env`
2. ตรวจสอบว่า PostgreSQL รันอยู่และเข้าถึงได้จาก server
3. ตรวจสอบ credentials และ network connectivity

---

## Security Checklist

### Production Deployment

- [ ] เปลี่ยน `JWT_SECRET` เป็นค่า random 32+ bytes
- [ ] ตั้งค่า `FRONTEND_ORIGIN` ให้เฉพาะเจาะจง (ไม่ใช่ `*`)
- [ ] ใช้ HTTPS สำหรับทุก connection
- [ ] เปิด firewall เฉพาะ port ที่จำเป็น
- [ ] ตั้งค่า rate limiting ที่ gateway
- [ ] ใช้ reverse proxy (nginx/Apache) ไม่ expose port โดยตรง
- [ ] ตั้งค่า `DB_MAX_OPEN` และ `DB_MAX_LIFETIME_MIN` ให้เหมาะสม
- [ ] Backup database เป็นประจำ
- [ ] Monitor logs และตั้ง alert

### JWT Secret Generation

```bash
# Generate random 32-byte secret (64 hex characters)
openssl rand -hex 32

# ตัวอย่าง output:
# a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2
```

ใส่ใน `backend/.env`:
```env
JWT_SECRET=a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2
```

---

## Performance Tuning

### Backend

```env
# จำนวน connection pool สูงสุด
DB_MAX_OPEN=50

# เวลาสูงสุดที่ connection จะถูก reuse (นาที)
DB_MAX_LIFETIME_MIN=60

# JWT token validity (นาที) - 8 ชั่วโมง = 480 นาที
JWT_TTL_MINUTES=480
```

### API Gateway

```env
# ถ้าต้องการเปลี่ยน port
GATEWAY_ADDR=0.0.0.0:8080

# ถ้า backend อยู่คนละเครื่อง
BACKEND_URL=http://10.0.32.71:12080
```

### Nginx

```nginx
# เพิ่ม worker processes
worker_processes auto;

# เพิ่ม keepalive connections
upstream sso_backend {
    server 127.0.0.1:18000;
    keepalive 32;
}

location /api/SSO_login/ {
    proxy_pass http://sso_backend/api/SSO_login/;
    proxy_http_version 1.1;
    proxy_set_header Connection "";
    # ... other settings
}
```

---

## Monitoring & Logging

### Logs Location

- **Backend:** stdout/stderr (หรือ systemd journal)
- **Gateway:** stdout/stderr (หรือ systemd journal)
- **Frontend:** Browser console (development)

### Viewing Logs

```bash
# systemd
sudo journalctl -u sso-backend -f
sudo journalctl -u sso-gateway -f

# Docker
docker logs -f sso-backend
docker logs -f sso-gateway
```

### Health Monitoring

ตั้ง endpoint monitoring ที่:
- `http://localhost:12080/health` - Backend
- `http://localhost:18000/api/SSO_login/health` - Gateway

ตัวอย่าง Prometheus metrics (ถ้าต้องการเพิ่ม):
```yaml
scrape_configs:
  - job_name: 'sso-login'
    static_configs:
      - targets: ['localhost:12080', 'localhost:18000']
```

---

## Version Information

- **Backend:** Go 1.23+
- **Frontend:** React 18+, Vite 5+
- **Database:** PostgreSQL 13+
- **Gateway:** Go 1.23+ (standard library)

---

## Support

หากมีปัญหาหรือข้อสงสัย:

1. อ่านเอกสารนี้และ [INTEGRATION_GUIDE.md](./INTEGRATION_GUIDE.md)
2. ตรวจสอบ logs ของแต่ละ service
3. ทดสอบ API ด้วย curl หรือ Postman
4. ติดต่อทีมพัฒนา SSO Login

---

**Version:** 1.0  
**Last Updated:** 2026-08-18
