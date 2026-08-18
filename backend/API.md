# API - sso-login backend

Base URL: `http://<host>:12010`

ทุก response เป็น JSON

---

## Auth (สำหรับหน้า management UI)

### POST /api/auth/login

```json
{ "username": "admin", "password": "any" }
```

Response 200:

```json
{
  "token": "5f1a8e9b...",
  "username": "admin",
  "displayName": "admin",
  "expiresAt": "2026-08-18T18:00:00Z"
}
```

### POST /api/auth/logout

ใช้ header `Authorization: Bearer <token>` → ตอบ `{"ok": true}`

### GET /api/auth/me

```json
{ "username": "admin" }
```

---

## POST /api/check-access

ใช้ตรวจสิทธิ์ผู้ใช้ที่ขอเข้าระบบ โดยดูจาก base URL + IP + AD username

### Request

```json
{
  "baseUrl": "http://10.0.32.71/HelpDesk/",
  "clientIp": "10.0.32.100",
  "adUsername": "somchai.s",
  "employeeId": "EMP001",
  "displayName": "สมชาย สามชื่อ",
  "userAgent": "Mozilla/5.0 ..."
}
```

* ถ้าไม่ระบุ `clientIp` server จะดึงจาก `X-Forwarded-For` หรือ `X-Real-IP` หรือ `RemoteAddr`
* ถ้าไม่ระบุ `userAgent` server จะใช้ header `User-Agent`

### Response 200 (ALLOW)

```json
{
  "allow": true,
  "result": "ALLOW",
  "appCode": "HELPDESK",
  "envCode": "PROD",
  "baseUrl": "http://10.0.32.71/HelpDesk/",
  "clientIp": "10.0.32.100",
  "adUsername": "somchai.s"
}
```

### Response 403 (DENY)

```json
{
  "allow": false,
  "result": "DENY_IP",
  "reason": "client_ip is not in allowed list",
  "baseUrl": "http://10.0.32.71/HelpDesk/",
  "clientIp": "10.115.2.50",
  "adUsername": "somchai.s"
}
```

`result` ที่เป็นไปได้:

| result      | ความหมาย                                         |
|-------------|---------------------------------------------------|
| `ALLOW`     | ผ่าน                                              |
| `DENY_APP`  | base_url ไม่ตรงกับ env ที่ลงทะเบียน                 |
| `DENY_IP`   | IP ไม่อยู่ใน allowed list                          |
| `DENY_USER` | AD username ไม่อยู่ใน allowed list หรือหมดอายุ    |
| `ERROR`     | server error                                       |

---

## Applications

### GET /api/apps

```json
{ "apps": [
  { "id": 1, "code": "HELPDESK", "name": "HelpDesk System", "description": "...", "active": true,
    "createdAt": "2026-08-18T10:00:00Z", "updatedAt": null }
]}
```

### POST /api/apps

```json
{ "code": "ERP", "name": "ERP System", "description": "ระบบ ERP" }
```

### PUT /api/apps/{id}

```json
{ "name": "ERP System v2", "description": "...", "active": true }
```

### DELETE /api/apps/{id}

ตอบ `200` เมื่อลบสำเร็จ

---

## Environments

### GET /api/envs?appId=1

```json
{ "envs": [
  { "id": 1, "appId": 1, "appCode": "HELPDESK", "envCode": "PROD", "envName": "Production",
    "baseUrl": "http://10.0.32.71/HelpDesk/", "hostIp": "10.0.32.71", "basePath": "/HelpDesk/",
    "adUser": "admin", "active": true, "createdAt": "...", "updatedAt": null }
]}
```

### POST /api/envs

```json
{ "appId": 1, "envCode": "UAT", "envName": "UAT",
  "baseUrl": "http://10.115.5.10/HelpDesk/", "hostIp": "10.115.5.10", "basePath": "/HelpDesk/",
  "adUser": "admin" }
```

### PUT /api/envs/{id}

```json
{ "envCode": "UAT", "envName": "UAT (Updated)",
  "baseUrl": "http://10.115.5.10/HelpDesk/", "hostIp": "10.115.5.10", "basePath": "/HelpDesk/",
  "adUser": "admin", "active": true }
```

> ต้องรัน migration `002_add_aduser_to_envs.sql` ก่อนใช้งาน field `adUser`

---

## Allowed IPs

### GET /api/allowed-ips?envId=1

```json
{ "allowedIps": [
  { "id": 1, "envId": 1, "ipCidr": "10.0.32.0/24", "description": "Office subnet",
    "active": true, "createdAt": "...", "createdBy": "admin" }
]}
```

### POST /api/allowed-ips

```json
{ "envId": 1, "ipCidr": "10.0.32.0/24", "description": "Office subnet", "createdBy": "admin" }
```

* `ipCidr` รองรับทั้ง IP เดี่ยว (`10.0.32.71`) และ CIDR (`10.0.32.0/24`)
* ใช้ `INET` ของ PostgreSQL ตรวจ `<<=` วง subnet

### DELETE /api/allowed-ips/{id}

---

## Allowed AD Users

### GET /api/allowed-users?envId=1

```json
{ "allowedUsers": [
  { "id": 1, "envId": 1, "adUsername": "somchai.s", "employeeId": "EMP001",
    "displayName": "สมชาย สามชื่อ", "email": "somchai@company.local", "department": "IT",
    "active": true, "grantedAt": "...", "grantedBy": "admin",
    "expiresAt": null, "lastSyncAt": null }
]}
```

### POST /api/allowed-users

```json
{ "envId": 1, "adUsername": "somchai.s", "employeeId": "EMP001",
  "displayName": "สมชาย สามชื่อ", "email": "somchai@company.local", "department": "IT",
  "grantedBy": "admin", "expiresAt": "2026-12-31T23:59:59Z" }
```

* `expiresAt` ไม่ระบุได้ = ไม่หมดอายุ

### DELETE /api/allowed-users/{id}

---

## Audit

### GET /api/audit?limit=100

```json
{ "audit": [
  { "id": 1, "appCode": "HELPDESK", "envCode": "PROD",
    "baseUrl": "http://10.0.32.71/HelpDesk/", "clientIp": "10.0.32.100",
    "adUsername": "somchai.s", "employeeId": "EMP001", "displayName": "...",
    "result": "ALLOW", "denyReason": "", "userAgent": "...",
    "requestId": "5f1a8e9b3c7d2f0a", "createdAt": "2026-08-18T10:05:30Z" }
]}
```
