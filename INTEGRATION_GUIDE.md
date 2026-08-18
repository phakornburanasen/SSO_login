# SSO Login Integration Guide

คู่มือสำหรับนักพัฒนาที่ต้องการให้แอปพลิเคชันอื่นใช้ระบบ SSO Login นี้ในการตรวจสอบสิทธิ์การเข้าถึง

---

## Architecture Overview

```
─────────────────────────────────────────────────────────────────┐
│                        Client Browser                           │
│  (แอปพลิเคชันอื่นที่ต้องการใช้ SSO)                              │
└──────────────────────────┬──────────────────────────────────────┘
                           │ HTTP/HTTPS
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                    API Gateway (Port 18000)                      │
│  - Reverse Proxy                                                │
│  - Strip prefix: /api/SSO_login                                 │
│  - Forward to Backend                                           │
│  - CORS Headers (จาก Backend)                                    │
└──────────────────────────┬──────────────────────────────────────┘
                           │ HTTP
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Backend API (Port 12080)                      │
│  - JWT Authentication                                           │
│  - Access Control Logic                                         │
│  - PostgreSQL Database                                          │
└─────────────────────────────────────────────────────────────────┘
```

---

## 1. การเข้าถึงระบบจาก IP อื่น

### ✅ สามารถเข้าถึงได้จากเครื่องอื่น

ระบบนี้ **รองรับการเข้าถึงจากเครื่องอื่นในเครือข่าย** โดยต้องตั้งค่าดังนี้:

#### 1.1 Backend Configuration

แก้ไขไฟล์ `backend/.env`:

```env
# เปลี่ยนจาก 127.0.0.1 เป็น 0.0.0.0 เพื่อให้รับ connection จากทุก IP
HTTP_ADDR=0.0.0.0:12080

# ตั้งค่า FRONTEND_ORIGIN ให้รองรับ origin ที่ต้องการ
# ใช้ * สำหรับอนุญาตทุก origin (เหมาะสำหรับ development)
# หรือระบุเฉพาะ origin ที่ต้องการ (แนะนำสำหรับ production)
FRONTEND_ORIGIN=*

# หรือระบุหลาย origins (คั่นด้วย comma)
# FRONTEND_ORIGIN=http://app1.example.com,http://app2.example.com
```

#### 1.2 API Gateway Configuration

แก้ไขไฟล์ `api_gatewayGo/.env`:

```env
# Gateway รับ connection จากทุก IP
GATEWAY_ADDR=0.0.0.0:18000

# Backend URL - ชี้ไปที่ backend (อาจอยู่คนละเครื่อง)
BACKEND_URL=http://10.0.32.71:12080
```

#### 1.3 Firewall Rules

ต้องเปิด port ใน firewall:
- **Port 18000** - API Gateway
- **Port 12080** - Backend (ถ้าต้องการให้เข้าถึงตรง)

```powershell
# Windows Firewall
New-NetFirewallRule -DisplayName "SSO Login Gateway" -Direction Inbound -LocalPort 18000 -Protocol TCP -Action Allow
New-NetFirewallRule -DisplayName "SSO Login Backend" -Direction Inbound -LocalPort 12080 -Protocol TCP -Action Allow
```

---

## 2. API Endpoints สำหรับแอปพลิเคชันอื่น

### 2.1 ตรวจสอบสิทธิ์การเข้าถึง (Access Check)

**Endpoint:** `POST /api/SSO_login/api/check-access`

**Request:**
```json
{
  "baseUrl": "http://10.0.32.71/HelpDesk/",
  "clientIp": "192.168.1.100",
  "adUsername": "somchai.s"
}
```

**Response (อนุญาต):**
```json
{
  "allow": true,
  "envId": 1,
  "envCode": "PROD",
  "appCode": "HELPDESK",
  "baseUrl": "http://10.0.32.71/HelpDesk/",
  "adUsername": "somchai.s",
  "employeeId": "E001",
  "displayName": "สมชาย ใจดี",
  "email": "somchai@example.com",
  "department": "IT"
}
```

**Response (ไม่อนุญาต):**
```json
{
  "allow": false,
  "denyReason": "IP 192.168.1.100 ไม่อยู่ใน allowed list ของ environment นี้"
}
```

**HTTP Status Codes:**
- `200 OK` - อนุญาต
- `403 Forbidden` - ไม่อนุญาต
- `404 Not Found` - ไม่พบ environment
- `500 Internal Server Error` - เกิดข้อผิดพลาด

---

### 2.2 ตัวอย่างการเรียกใช้จากแอปพลิเคชันอื่น

#### JavaScript / TypeScript (Browser)

```javascript
const API_BASE = 'http://10.0.32.71:18000/api/SSO_login';

async function checkAccess(baseUrl, clientIp, adUsername) {
  try {
    const response = await fetch(`${API_BASE}/api/check-access`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        baseUrl,
        clientIp,
        adUsername,
      }),
    });

    const data = await response.json();

    if (response.status === 200) {
      // อนุญาต - data.allow === true
      console.log('Access granted:', data);
      return { allowed: true, userInfo: data };
    } else if (response.status === 403) {
      // ไม่อนุญาต
      console.log('Access denied:', data.denyReason);
      return { allowed: false, reason: data.denyReason };
    } else {
      // Error อื่นๆ
      console.error('Error:', data);
      return { allowed: false, reason: 'เกิดข้อผิดพลาดในการตรวจสอบ' };
    }
  } catch (error) {
    console.error('Network error:', error);
    return { allowed: false, reason: 'ไม่สามารถเชื่อมต่อระบบได้' };
  }
}

// วิธีใช้
const result = await checkAccess(
  'http://10.0.32.71/HelpDesk/',
  '192.168.1.100',
  'somchai.s'
);

if (result.allowed) {
  // นำทางไปยังระบบ
  window.location.href = result.userInfo.baseUrl;
} else {
  alert('คุณไม่มีสิทธิ์เข้าถึงระบบนี้: ' + result.reason);
}
```

#### Python (Backend-to-Backend)

```python
import requests

API_BASE = 'http://10.0.32.71:18000/api/SSO_login'

def check_access(base_url: str, client_ip: str, ad_username: str) -> dict:
    """
    ตรวจสอบสิทธิ์การเข้าถึงระบบ
    
    Returns:
        dict: {
            'allowed': bool,
            'user_info': dict | None,
            'reason': str | None
        }
    """
    try:
        response = requests.post(
            f'{API_BASE}/api/check-access',
            json={
                'baseUrl': base_url,
                'clientIp': client_ip,
                'adUsername': ad_username,
            },
            timeout=5,
        )
        
        if response.status_code == 200:
            data = response.json()
            return {
                'allowed': True,
                'user_info': data,
                'reason': None,
            }
        elif response.status_code == 403:
            data = response.json()
            return {
                'allowed': False,
                'user_info': None,
                'reason': data.get('denyReason', 'ไม่มีสิทธิ์เข้าถึง'),
            }
        else:
            return {
                'allowed': False,
                'user_info': None,
                'reason': f'HTTP {response.status_code}',
            }
    except requests.exceptions.RequestException as e:
        return {
            'allowed': False,
            'user_info': None,
            'reason': f'Connection error: {str(e)}',
        }

# วิธีใช้
result = check_access(
    base_url='http://10.0.32.71/HelpDesk/',
    client_ip='192.168.1.100',
    ad_username='somchai.s',
)

if result['allowed']:
    print(f"อนุญาตให้ {result['user_info']['displayName']} เข้าถึงระบบ")
else:
    print(f"ปฏิเสธ: {result['reason']}")
```

#### C# / .NET

```csharp
using System.Net.Http;
using System.Text;
using System.Text.Json;

public class SsoAccessChecker
{
    private readonly HttpClient _httpClient;
    private readonly string _apiBase;

    public SsoAccessChecker(string apiBase)
    {
        _apiBase = apiBase.TrimEnd('/');
        _httpClient = new HttpClient();
    }

    public async Task<AccessResult> CheckAccessAsync(
        string baseUrl, 
        string clientIp, 
        string adUsername)
    {
        try
        {
            var payload = new
            {
                baseUrl,
                clientIp,
                adUsername
            };

            var json = JsonSerializer.Serialize(payload);
            var content = new StringContent(json, Encoding.UTF8, "application/json");

            var response = await _httpClient.PostAsync(
                $"{_apiBase}/api/check-access", 
                content);

            var responseJson = await response.Content.ReadAsStringAsync();

            if (response.StatusCode == System.Net.HttpStatusCode.OK)
            {
                var data = JsonSerializer.Deserialize<JsonElement>(responseJson);
                return new AccessResult
                {
                    Allowed = true,
                    UserInfo = data,
                    Reason = null
                };
            }
            else if (response.StatusCode == System.Net.HttpStatusCode.Forbidden)
            {
                var data = JsonSerializer.Deserialize<JsonElement>(responseJson);
                return new AccessResult
                {
                    Allowed = false,
                    UserInfo = null,
                    Reason = data.GetProperty("denyReason").GetString()
                };
            }
            else
            {
                return new AccessResult
                {
                    Allowed = false,
                    UserInfo = null,
                    Reason = $"HTTP {(int)response.StatusCode}"
                };
            }
        }
        catch (Exception ex)
        {
            return new AccessResult
            {
                Allowed = false,
                UserInfo = null,
                Reason = $"Error: {ex.Message}"
            };
        }
    }
}

public class AccessResult
{
    public bool Allowed { get; set; }
    public JsonElement? UserInfo { get; set; }
    public string Reason { get; set; }
}

// วิธีใช้
var checker = new SsoAccessChecker("http://10.0.32.71:18000/api/SSO_login");
var result = await checker.CheckAccessAsync(
    "http://10.0.32.71/HelpDesk/",
    "192.168.1.100",
    "somchai.s"
);

if (result.Allowed)
{
    Console.WriteLine($"อนุญาตให้เข้าถึงระบบ");
}
else
{
    Console.WriteLine($"ปฏิเสธ: {result.Reason}");
}
```

---

## 3. การตั้งค่า CORS สำหรับเครื่องอื่น

### 3.1 Backend CORS Configuration

Backend ใช้ `FRONTEND_ORIGIN` ใน `.env` เพื่อควบคุม CORS:

```env
# อนุญาตทุก origin (เหมาะสำหรับ development)
FRONTEND_ORIGIN=*

# หรือระบุเฉพาะ origin (แนะนำสำหรับ production)
FRONTEND_ORIGIN=http://app1.example.com,http://app2.example.com
```

**ตัวอย่าง:** ถ้าแอปพลิเคชันอื่นอยู่ที่ `http://10.0.32.72:3000`

```env
FRONTEND_ORIGIN=http://10.0.32.72:3000
```

### 3.2 API Gateway CORS

Gateway **ไม่เพิ่ม CORS headers ซ้ำ** สำหรับ proxied paths - ปล่อยให้ Backend จัดการ

แต่สำหรับ `/health` endpoint ของ Gateway เอง จะมี CORS middleware ที่อนุญาตทุก origin

### 3.3 ทดสอบ CORS

```bash
# ทดสอบจากเครื่องอื่น
curl -X POST \
  -H "Content-Type: application/json" \
  -H "Origin: http://10.0.32.72:3000" \
  -d '{"baseUrl":"http://10.0.32.71/HelpDesk/","clientIp":"192.168.1.100","adUsername":"somchai.s"}' \
  http://10.0.32.71:18000/api/SSO_login/api/check-access
```

ตรวจสอบ response headers:
```
Access-Control-Allow-Origin: http://10.0.32.72:3000
Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
```

---

## 4. การตั้งค่า Frontend API Base

### 4.1 Dynamic Hostname (Default)

ระบบใช้ **dynamic hostname** จาก browser โดยอัตโนมัติ:

```javascript
// frontend/src/api.js
const GATEWAY_PORT = import.meta.env.VITE_GATEWAY_PORT || '18000'
const RAW_BASE = import.meta.env.VITE_API_BASE || `http://${window.location.hostname}:${GATEWAY_PORT}/api/SSO_login`
```

**วิธีทำงาน:**
- เปิดจากเครื่องไหน → เรียก gateway เครื่องนั้นอัตโนมัติ
- ไม่ต้องตั้งค่า `.env` สำหรับแต่ละเครื่อง
- ไม่ต้อง build ใหม่เมื่อ IP เปลี่ยน

**ตัวอย่าง:**
- เปิดจาก `http://10.0.32.71:3000` → เรียก `http://10.0.32.71:18000/api/SSO_login`
- เปิดจาก `http://192.168.1.100:3000` → เรียก `http://192.168.1.100:18000/api/SSO_login`

### 4.2 Override ด้วย Environment Variable

ถ้าต้องการใช้ URL เฉพาะ (เช่น production domain):

สร้างไฟล์ `frontend/.env`:

```env
# ใช้ domain เฉพาะ
VITE_API_BASE=https://sso.example.com/api/SSO_login

# หรือใช้ IP เฉพาะ
VITE_API_BASE=http://10.0.32.71:18000/api/SSO_login
```

### 4.3 เปลี่ยน Port Gateway

ถ้า gateway ใช้ port อื่นที่ไม่ใช่ 18000:

```env
# frontend/.env
VITE_GATEWAY_PORT=8080
```

หรือตั้งพร้อมกับ `VITE_API_BASE`:

```env
VITE_API_BASE=http://10.0.32.71:8080/api/SSO_login
```

---

## 5. Security Best Practices

### 5.1 Production Checklist

- [ ] เปลี่ยน `JWT_SECRET` ใน `backend/.env` เป็นค่า random 32+ bytes
- [ ] ตั้งค่า `FRONTEND_ORIGIN` ให้เฉพาะเจาะจง (ไม่ใช่ `*`)
- [ ] ใช้ HTTPS สำหรับทุก connection
- [ ] เปิด firewall เฉพาะ port ที่จำเป็น
- [ ] ตั้งค่า `DB_MAX_OPEN` และ `DB_MAX_LIFETIME_MIN` ให้เหมาะสม
- [ ] ใช้ reverse proxy (nginx/Apache) สำหรับ production
- [ ] ตั้งค่า rate limiting ที่ gateway

### 5.2 JWT Secret Generation

```bash
# Generate random 32-byte secret
openssl rand -hex 32
# Example output: a1b2c3d4e5f6... (64 hex characters)
```

ใส่ใน `backend/.env`:
```env
JWT_SECRET=a1b2c3d4e5f6...
```

---

## 6. Troubleshooting

### 6.1 CORS Error ใน Browser

**อาการ:** `Access to fetch at '...' has been blocked by CORS policy`

**วิธีแก้:**
1. ตรวจสอบ `FRONTEND_ORIGIN` ใน `backend/.env`
2. ถ้าใช้ `*` ให้แน่ใจว่าไม่ได้ส่ง credentials
3. ตรวจสอบว่า Gateway ไม่ได้เพิ่ม CORS headers ซ้ำ

### 6.2 Connection Refused

**อาการ:** `Unable to connect to API Gateway`

**วิธีแก้:**
1. ตรวจสอบว่า Gateway รันอยู่: `netstat -an | findstr 18000`
2. ตรวจสอบ firewall rules
3. ตรวจสอบ `BACKEND_URL` ใน `api_gatewayGo/.env`

### 6.3 401 Unauthorized

**อาการ:** `invalid or expired session`

**วิธีแก้:**
1. ตรวจสอบว่า JWT token ถูกส่งใน header: `Authorization: Bearer <token>`
2. ตรวจสอบว่า token ยังไม่หมดอายุ (default 8 ชั่วโมง)
3. Login ใหม่เพื่อรับ token ใหม่

---

## 7. ตัวอย่าง Integration แบบเต็ม

### React Application

```jsx
import { useState, useEffect } from 'react';

const API_BASE = import.meta.env.VITE_API_BASE || '/api/SSO_login';

function App() {
  const [access, setAccess] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    checkMyAccess();
  }, []);

  const checkMyAccess = async () => {
    try {
      // สมมติว่าได้ adUsername จาก authentication อื่น
      const adUsername = getCurrentUser(); // implement ตามระบบของคุณ
      const clientIp = await getClientIp(); // implement ตามระบบของคุณ

      const response = await fetch(`${API_BASE}/api/check-access`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          baseUrl: window.location.origin, // ระบบปัจจุบัน
          clientIp,
          adUsername,
        }),
      });

      if (response.status === 200) {
        const data = await response.json();
        setAccess({ allowed: true, userInfo: data });
      } else {
        setAccess({ allowed: false, reason: 'ไม่มีสิทธิ์เข้าถึง' });
      }
    } catch (error) {
      setAccess({ allowed: false, reason: 'ไม่สามารถเชื่อมต่อระบบได้' });
    } finally {
      setLoading(false);
    }
  };

  if (loading) return <div>กำลังตรวจสอบสิทธิ์...</div>;

  if (!access.allowed) {
    return (
      <div className="access-denied">
        <h1>ไม่มีสิทธิ์เข้าถึง</h1>
        <p>{access.reason}</p>
        <p>กรุณาติดต่อผู้ดูแลระบบเพื่อขอสิทธิ์</p>
      </div>
    );
  }

  return (
    <div className="app">
      <h1>ยินดีต้อนรับ {access.userInfo.displayName}</h1>
      {/* เนื้อหาแอปพลิเคชัน */}
    </div>
  );
}

export default App;
```

---

## 8. API Reference

### POST /api/SSO_login/api/check-access

ตรวจสอบสิทธิ์การเข้าถึงระบบ

**Request Body:**
```json
{
  "baseUrl": "string (required) - URL ของระบบที่ต้องการเข้าถึง",
  "clientIp": "string (required) - IP ของ client",
  "adUsername": "string (required) - AD username",
  "userAgent": "string (optional) - User agent string"
}
```

**Response 200 (อนุญาต):**
```json
{
  "allow": true,
  "envId": 1,
  "envCode": "PROD",
  "appCode": "HELPDESK",
  "baseUrl": "http://10.0.32.71/HelpDesk/",
  "adUsername": "somchai.s",
  "employeeId": "E001",
  "displayName": "สมชาย ใจดี",
  "email": "somchai@example.com",
  "department": "IT"
}
```

**Response 403 (ไม่อนุญาต):**
```json
{
  "allow": false,
  "denyReason": "เหตุผลที่ปฏิเสธ"
}
```

---

## 9. ติดต่อและสนับสนุน

หากมีปัญหาหรือข้อสงสัยในการ integrate:

1. ตรวจสอบเอกสารนี้
2. ดู logs ของ Backend และ Gateway
3. ทดสอบ API ด้วย curl หรือ Postman
4. ติดต่อทีมพัฒนา SSO Login

---

**Version:** 1.0  
**Last Updated:** 2026-08-18
