# SSO Login - Quick Reference Card

สำหรับนักพัฒนาที่ต้องการ integrate กับระบบ SSO Login อย่างรวดเร็ว

---

## 🔑 สิ่งที่ต้องรู้

### Base URL
```
Default:      http://<hostname ของเครื่องที่เปิด>:18000/api/SSO_login
Override:     https://sso.example.com/api/SSO_login  (ตั้งค่าใน .env)
```

### Endpoint สำคัญ
```
POST /api/check-access  ← ใช้บ่อยที่สุด
```

---

## 🚀 Quick Start (5 นาที)

### Step 1: ตรวจสอบสิทธิ์

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
  // ✅ อนุญาต - เข้าถึงระบบได้
  console.log('Welcome:', data.displayName);
} else {
  //  ไม่อนุญาต
  console.log('Denied:', data.denyReason);
}
```

### Step 2: จัดการผลลัพธ์

```javascript
// Response เมื่ออนุญาต (200)
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

// Response เมื่อไม่อนุญาต (403)
{
  "allow": false,
  "denyReason": "IP ไม่อยู่ใน allowed list"
}
```

---

## 📋 HTTP Status Codes

| Code | ความหมาย | การจัดการ |
|------|----------|-----------|
| 200  | อนุญาต | เข้าถึงระบบได้ |
| 403  | ไม่อนุญาต | แสดงข้อความปฏิเสธ |
| 404  | ไม่พบระบบ | ตรวจสอบ baseUrl |
| 500  | Error | ลองใหม่หรือติดต่อ admin |

---

## 🔧 ตัวอย่างตามภาษา

### Python
```python
import requests

response = requests.post(
    'http://10.0.32.71:18000/api/SSO_login/api/check-access',
    json={
        'baseUrl': 'http://10.0.32.71/HelpDesk/',
        'clientIp': '192.168.1.100',
        'adUsername': 'somchai.s'
    }
)

if response.status_code == 200:
    data = response.json()
    print(f"อนุญาต: {data['displayName']}")
else:
    print(f"ปฏิเสธ: {response.json()['denyReason']}")
```

### C# / .NET
```csharp
var client = new HttpClient();
var response = await client.PostAsJsonAsync(
    "http://10.0.32.71:18000/api/SSO_login/api/check-access",
    new {
        baseUrl = "http://10.0.32.71/HelpDesk/",
        clientIp = "192.168.1.100",
        adUsername = "somchai.s"
    }
);

if (response.IsSuccessStatusCode) {
    var data = await response.Content.ReadFromJsonAsync<AccessResult>();
    Console.WriteLine($"อนุญาต: {data.DisplayName}");
}
```

### PHP
```php
$ch = curl_init('http://10.0.32.71:18000/api/SSO_login/api/check-access');
curl_setopt($ch, CURLOPT_POST, true);
curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode([
    'baseUrl' => 'http://10.0.32.71/HelpDesk/',
    'clientIp' => '192.168.1.100',
    'adUsername' => 'somchai.s'
]));
curl_setopt($ch, CURLOPT_HTTPHEADER, ['Content-Type: application/json']);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);

$response = curl_exec($ch);
$data = json_decode($response, true);

if (curl_getinfo($ch, CURLINFO_HTTP_CODE) == 200 && $data['allow']) {
    echo "อนุญาต: " . $data['displayName'];
} else {
    echo "ปฏิเสธ: " . $data['denyReason'];
}
```

---

## 🌐 CORS Configuration

### ถ้าเจอ CORS Error

**Backend (.env):**
```env
# อนุญาตทุก origin (development)
FRONTEND_ORIGIN=*

# หรือระบุเฉพาะ (production)
FRONTEND_ORIGIN=https://app.example.com
```

**Frontend (.env) — ถ้าต้องการ override:**
```env
# ใช้ URL เฉพาะแทน dynamic hostname
VITE_API_BASE=https://sso.example.com/api/SSO_login
```

**หมายเหตุ:** โดย default ระบบใช้ dynamic hostname จาก browser อัตโนมัติ
- เปิดจาก `http://10.0.32.71:3000` → เรียก `http://10.0.32.71:18000/api/SSO_login`
- เปิดจาก `http://192.168.1.100:3000` → เรียก `http://192.168.1.100:18000/api/SSO_login`

---

## 🧪 ทดสอบด้วย curl

```bash
# ทดสอบ health
curl http://10.0.32.71:18000/api/SSO_login/health

# ทดสอบ check-access
curl -X POST http://10.0.32.71:18000/api/SSO_login/api/check-access \
  -H "Content-Type: application/json" \
  -d '{
    "baseUrl": "http://10.0.32.71/HelpDesk/",
    "clientIp": "192.168.1.100",
    "adUsername": "somchai.s"
  }'
```

---

## 📚 เอกสารเพิ่มเติม

- [INTEGRATION_GUIDE.md](./INTEGRATION_GUIDE.md) - คู่มือ integration แบบละเอียด
- [DEPLOYMENT.md](./DEPLOYMENT.md) - คู่มือ deployment
- [README.md](./README.md) - ภาพรวมโปรเจค

---

##  Tips

1. **Dynamic Hostname** - ระบบใช้ hostname จาก browser อัตโนมัติ ไม่ต้องตั้งค่า .env
2. **Cache Results** - ผลลัพธ์ check-access สามารถ cache ได้ 5-10 นาทีเพื่อลด load
3. **Handle Errors** - เสมอ handle network errors และ timeout
4. **Log Everything** - บันทึก log ทั้ง allowed และ denied requests
5. **Test First** - ทดสอบด้วย curl ก่อน integrate จริง
6. **Override ถ้าต้องการ** - ตั้งค่า `VITE_API_BASE` ใน .env ถ้าต้องการใช้ URL เฉพาะ

---

**Version:** 1.0  
**Last Updated:** 2026-08-18
