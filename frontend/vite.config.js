import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// base path ของ frontend เมื่อเสิร์ฟผ่าน web server (nginx/IIS) ด้านหน้า api_gatewayGo
// เช่น https://gateway/SSO_login/  ->  index.html จะโหลดจาก /SSO_login/
//
// API: เรียกตรงไปที่ gateway port 18000 (ดูค่า default ใน src/api.js)
// ตอน dev ไม่ต้องใช้ proxy — ให้เห็นใน Network tab ว่าไปที่ :18000 ชัด ๆ
// (port 6000 โดน Chrome block เพราะเป็น X11 — ERR_UNSAFE_PORT)
export default defineConfig({
  base: '/SSO_login/',
  plugins: [react()],
  server: {
    host: true,
    port: 3000,
    strictPort: true,
  },
  preview: {
    host: true,
    port: 3000,
    strictPort: true,
  },
})
