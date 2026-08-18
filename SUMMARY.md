# สรุปการแก้ไขและเอกสารประกอบ

## สิ่งที่ทำในรอบนี้

### 1. แก้ Bug `invalid json: parsing time` ✅

**ปัญหา:** เมื่อส่ง `expiresAt` จาก frontend ใน format `2030-12-18T20:40` (จาก `<input type="datetime-local">`) backend ไม่สามารถ parse ได้

**วิธีแก้:** สร้าง `FlexibleTime` type ใน [model.go](file:///d:/Dev/SSO_login/backend/internal/model/model.go) ที่รองรับหลาย format:
- `2006-01-02T15:04:05Z07:00` (RFC3339)
- `2006-01-02T15:04` (datetime-local จาก browser)
- `2006-01-02` (date only)

---

### 2. ปรับหน้า Allowed User ใหม่ ✅

**ไฟล์:** [AllowedUserManager.jsx](file:///d:/Dev/SSO_login/frontend/src/components/AllowedUserManager.jsx)

**ฟีเจอร์ใหม่:**
- ✅ ใส่ AD Username แล้วเลือกระบบที่ต้องการให้สิทธิ์แบบ checkbox
- ✅ ระบบแสดงเป็นกลุ่มตาม Application (AppCode)
- ✅ เลือกทั้งหมด / ล้างทั้งหมด ได้ในคลิกเดียว
- ✅ **อ้างอิงสิทธิ์จาก AD User อื่น** - ดึงสิทธิ์ของผู้นั้นมาแสดงและติ๊กให้อัตโนมัติ
- ✅ เลือกวันหมดอายุ (ไม่บังคับ)
- ✅ บันทึกทีเดียวหลายระบบผ่าน API `/api/allowed-users/bulk`
- ✅ แสดงตารางสิทธิ์ที่มีอยู่ พร้อมกรองตาม Environment

**Backend APIs ใหม่:**
- `POST /api/allowed-users/bulk` - สร้าง allowed users หลาย env พร้อมกัน
- `GET /api/allowed-users/by-user?username=xxx` - ดูสิทธิ์ของ AD user คนหนึ่ง

---

### 3. Sidebar ซ่อน/แสดงได้ + Persist State ✅

**ไฟล์:** [Dashboard.jsx](file:///d:/Dev/SSO_login/frontend/src/pages/Dashboard.jsx)

- ✅ เพิ่มปุ่ม toggle (hamburger icon) ใน navbar
- ✅ Sidebar transition นุ่มนวล (300ms)
- ✅ Persist state ใน `localStorage` (key: `sso_sidebar_open`)
- ✅ รีเฟรชหน้าเว็บแล้ว Sidebar ไม่เด้งเอง

---

### 4. ปรับหน้า Login ✅

**ไฟล์:** [Login.jsx](file:///d:/Dev/SSO_login/frontend/src/pages/Login.jsx)

- ✅ พื้นหลัง gradient อ่อนๆ (slate-50 → white → blue-50/30)
- ✅ Card สีขาว rounded-2xl พร้อม shadow นุ่มนวล
- ✅ Logo icon แบบ gradient brand color
- ✅ ฟอร์มเรียบง่าย ใช้ `.input` class เดียวกับระบบ
- ✅ ปุ่มเข้าสู่ระบบใช้ `.btn-primary` แบบเต็มความกว้าง

---

### 5. UI สไตล์ TNL Guest Wi-Fi ✅

ปรับทั้งระบบให้:
- ✅ **Rounded corners** (rounded-xl, rounded-2xl)
- ✅ **Soft shadows** - เงาเบาๆ ไม่จัดจ้าน
- ✅ **Clean borders** - border-slate-200/80 จางๆ
- ✅ **Smooth transitions** - ทุก interaction มี transition นุ่มนวล
- ✅ **Fade-in animations** - ข้อมูลโหลดเสร็จค่อยๆ ปรากฏ
- ✅ **Modal slide-up** - Modal เลื่อนขึ้นมานุ่มนวล

---

### 6. แก้ RAW_BASE ให้ใช้ Dynamic Hostname ✅

**ไฟล์:** [api.js](file:///d:/Dev/SSO_login/frontend/src/api.js)

**ก่อน:**
```javascript
const RAW_BASE = import.meta.env.VITE_API_BASE || 'http://172.20.10.2:18000/api/SSO_login'
```

**หลัง:**
```javascript
const GATEWAY_PORT = import.meta.env.VITE_GATEWAY_PORT || '18000'
const RAW_BASE = import.meta.env.VITE_API_BASE || `http://${window.location.hostname}:${GATEWAY_PORT}/api/SSO_login`
```

**ข้อดี:**
- ✅ เปิดจากเครื่องไหนก็เรียก gateway เครื่องนั้นอัตโนมัติ
- ✅ ไม่ต้องตั้งค่า `.env` สำหรับแต่ละเครื่อง
- ✅ ไม่ต้อง build ใหม่เมื่อ IP เปลี่ยน
- ✅ ถ้าต้องการใช้ URL เฉพาะ ตั้งค่า `VITE_API_BASE` ใน `.env` ได้

---

### 7. สร้างเอกสารประกอบ ✅

สร้างเอกสาร 3 ไฟล์:

1. **[INTEGRATION_GUIDE.md](file:///d:/Dev/SSO_login/INTEGRATION_GUIDE.md)** - คู่มือสำหรับนักพัฒนาที่ต้องการให้แอปพลิเคชันอื่นใช้ระบบ SSO Login
   - Architecture overview
   - การเข้าถึงระบบจาก IP อื่น
   - API endpoints และตัวอย่าง code (JavaScript, Python, C#)
   - CORS configuration
   - Security best practices
   - Troubleshooting

2. **[DEPLOYMENT.md](file:///d:/Dev/SSO_login/DEPLOYMENT.md)** - คู่มือ deployment สำหรับสภาพแวดล้อมต่างๆ
   - Scenario 1: Single Server (แนะนำ)
   - Scenario 2: Separate Servers
   - Scenario 3: Development Environment
   - Firewall configuration
   - systemd services
   - Nginx configuration
   - Performance tuning
   - Monitoring & logging

3. **[QUICK_START.md](file:///d:/Dev/SSO_login/QUICK_START.md)** - Quick reference card สำหรับ integrate อย่างรวดเร็ว
   - 5-minute quick start
   - ตัวอย่าง code ตามภาษา (JavaScript, Python, C#, PHP)
   - HTTP status codes
   - CORS troubleshooting
   - Tips & best practices

---

## คำตอบสำหรับคำถามของคุณ

### ❓ ถ้าแอปอื่นจะมาเรียกใช้งาน API login ต้องทำอย่างไร?

**ตอบ:** แอปอื่นไม่ต้อง login เข้าระบบ SSO โดยตรง แต่ใช้ **endpoint `/api/check-access`** เพื่อตรวจสอบสิทธิ์

**ตัวอย่าง:**
```javascript
const response = await fetch('http://10.0.32.71:18000/api/SSO_login/api/check-access', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    baseUrl: 'http://10.0.32.71/HelpDesk/',  // ระบบของคุณ
    clientIp: '192.168.1.100',                // IP ผู้ใช้
    adUsername: 'somchai.s'                   // AD username
  })
});

const data = await response.json();

if (response.status === 200 && data.allow) {
  // อนุญาต - เข้าถึงระบบได้
  window.location.href = data.baseUrl;
} else {
  // ไม่อนุญาต
  alert('คุณไม่มีสิทธิ์: ' + data.denyReason);
}
```

**ระบบจะตรวจสอบ:**
1. ✅ `baseUrl` ตรงกับ system ที่ลงทะเบียนไว้ไหม?
2. ✅ `clientIp` อยู่ใน allowed IPs ของ environment นั้นไหม?
3. ✅ `adUsername` อยู่ใน allowed users ของ environment นั้นไหม?
4. ✅ สิทธิ์ยังไม่หมดอายุใช่ไหม?

ถ้าผ่านทุกข้อ → **อนุญาต** (200)  
ถ้าไม่ผ่านข้อใดข้อหนึ่ง → **ปฏิเสธ** (403)

---

### ❓ ระบบนี้สามารถเข้าผ่าน IP อื่นได้ไหม?

**ตอบ:** ✅ **ได้** แต่ต้องตั้งค่า

**ขั้นตอน:**

1. **Backend** - เปลี่ยน `HTTP_ADDR` ให้รับจากทุก IP:
   ```env
   # backend/.env
   HTTP_ADDR=0.0.0.0:12080
   FRONTEND_ORIGIN=*  # หรือระบุเฉพาะ origin
   ```

2. **Gateway** - เปลี่ยน `GATEWAY_ADDR` ให้รับจากทุก IP:
   ```env
   # api_gatewayGo/.env
   GATEWAY_ADDR=0.0.0.0:18000
   BACKEND_URL=http://10.0.32.71:12080
   ```

3. **Firewall** - เปิด port 18000 และ 12080

4. **Frontend** - ตั้งค่า `VITE_API_BASE` ให้ชี้ไปที่ gateway IP:
   ```env
   # frontend/.env
   VITE_API_BASE=http://10.0.32.71:18000/api/SSO_login
   ```

ดูรายละเอียดใน [DEPLOYMENT.md](file:///d:/Dev/SSO_login/DEPLOYMENT.md) Scenario 2

---

### ❓ RAW_BASE ต้องเขียนอย่างไรถึงไม่ต้องเปลี่ยน IP?

**ตอบ:** ใช้ **dynamic hostname** จาก browser (แก้แล้วใน [api.js](file:///d:/Dev/SSO_login/frontend/src/api.js))

```javascript
const GATEWAY_PORT = import.meta.env.VITE_GATEWAY_PORT || '18000'
const RAW_BASE = import.meta.env.VITE_API_BASE || `http://${window.location.hostname}:${GATEWAY_PORT}/api/SSO_login`
```

**วิธีทำงาน:**

1. **Default (ไม่ต้องตั้งค่า):** ใช้ hostname + port 18000 อัตโนมัติ
   - เปิดจาก `http://10.0.32.71:3000` → เรียก `http://10.0.32.71:18000/api/SSO_login`
   - เปิดจาก `http://192.168.1.100:3000` → เรียก `http://192.168.1.100:18000/api/SSO_login`
   - ไม่ต้องตั้งค่า .env สำหรับแต่ละเครื่อง

2. **Override (ถ้าต้องการ):** ตั้งค่าผ่าน environment variable
   ```env
   # frontend/.env
   VITE_API_BASE=https://sso.example.com/api/SSO_login
   ```

**ข้อดี:**
- ✅ เปิดจากเครื่องไหนก็เรียก gateway เครื่องนั้นอัตโนมัติ
- ✅ ไม่ต้องตั้งค่า `.env` สำหรับแต่ละเครื่อง
- ✅ ไม่ต้อง build ใหม่เมื่อ IP เปลี่ยน
- ✅ ถ้าต้องการใช้ URL เฉพาะ ตั้งค่า `VITE_API_BASE` ใน `.env` ได้

---

### ❓ เครื่องอื่นๆ สามารถใช้ระบบได้ไม่ติดเรื่อง CORS API จาก backend?

**ตอบ:** ✅ **ได้** ถ้าตั้งค่าถูกต้อง

**CORS Configuration:**

Backend ใช้ `FRONTEND_ORIGIN` ใน `.env` เพื่อควบคุม CORS:

```env
# อนุญาตทุก origin (development)
FRONTEND_ORIGIN=*

# หรือระบุเฉพาะ origin (production - แนะนำ)
FRONTEND_ORIGIN=https://app1.example.com,https://app2.example.com
```

**Gateway:**
- Gateway **ไม่เพิ่ม CORS headers ซ้ำ** สำหรับ proxied paths
- ปล่อยให้ Backend จัดการ CORS headers
- สำหรับ `/health` endpoint ของ Gateway เอง จะมี CORS middleware ที่อนุญาตทุก origin

**ทดสอบ CORS:**
```bash
curl -I -X OPTIONS \
  -H "Origin: http://10.0.32.72:3000" \
  -H "Access-Control-Request-Method: POST" \
  http://10.0.32.71:18000/api/SSO_login/api/check-access
```

**Expected headers:**
```
Access-Control-Allow-Origin: http://10.0.32.72:3000
Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
```

ดูรายละเอียดใน [INTEGRATION_GUIDE.md](file:///d:/Dev/SSO_login/INTEGRATION_GUIDE.md) Section 3

---

## Build Status

✅ **Backend:** Build สำเร็จ  
✅ **Frontend:** Build สำเร็จ  
✅ **Gateway:** Build สำเร็จ  

---

## วิธีทดสอบ

1. **รันระบบ:**
   ```bash
   .\runserver.bat
   ```

2. **เข้าหน้า Login:**
   ```
   http://127.0.0.1:3000
   ```

3. **ทดสอบ Allowed User:**
   - เข้าหน้า "Allowed Users"
   - ใส่ AD Username เช่น `testuser`
   - กด "ดึงสิทธิ์" จาก user อื่น (ถ้ามี) หรือติ๊ก checkbox ระบบเอง
   - เลือกวันหมดอายุ (ถ้าต้องการ)
   - กด "บันทึกสิทธิ์"

4. **ทดสอบ Sidebar:**
   - กดปุ่ม hamburger ใน navbar เพื่อซ่อน/แสดงเมนู
   - รีเฟรชหน้าเว็บ - Sidebar จะคงสถานะเดิม

5. **ทดสอบ API จากเครื่องอื่น:**
   ```bash
   curl -X POST http://10.0.32.71:18000/api/SSO_login/api/check-access \
     -H "Content-Type: application/json" \
     -d '{
       "baseUrl": "http://10.0.32.71/HelpDesk/",
       "clientIp": "192.168.1.100",
       "adUsername": "somchai.s"
     }'
   ```

---

## เอกสารที่เกี่ยวข้อง

- [INTEGRATION_GUIDE.md](file:///d:/Dev/SSO_login/INTEGRATION_GUIDE.md) - คู่มือ integration แบบละเอียด
- [DEPLOYMENT.md](file:///d:/Dev/SSO_login/DEPLOYMENT.md) - คู่มือ deployment
- [QUICK_START.md](file:///d:/Dev/SSO_login/QUICK_START.md) - Quick reference card
- [README.md](file:///d:/Dev/SSO_login/README.md) - ภาพรวมโปรเจค

---

**Version:** 1.0  
**Last Updated:** 2026-08-18
