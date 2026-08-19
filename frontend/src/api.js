// API client — ยิงตรงไปที่ api_gatewayGo
// 
// Default: ใช้ dynamic hostname จาก browser + port 18000
//   - เปิดจากเครื่องไหนก็เรียก gateway เครื่องนั้นอัตโนมัติ
//   - ไม่ต้องตั้งค่า .env สำหรับแต่ละเครื่อง
//
// Override: ตั้งค่า VITE_API_BASE ใน .env ถ้าต้องการใช้ URL เฉพาะ
//   - ตัวอย่าง: VITE_API_BASE=https://sso.example.com/api/SSO_login

const GATEWAY_PORT = import.meta.env.VITE_GATEWAY_PORT || '18000'
const RAW_BASE = import.meta.env.VITE_API_BASE || `http://${window.location.hostname}:${GATEWAY_PORT}/api/SSO_login`
export const API_BASE = RAW_BASE.replace(/\/+$/, '')

const TOKEN_KEY = 'sso_login_token'
const USER_KEY  = 'sso_login_user'

export function getToken() {
  return localStorage.getItem(TOKEN_KEY) || ''
}
export function setToken(t) {
  if (t) localStorage.setItem(TOKEN_KEY, t)
  else   localStorage.removeItem(TOKEN_KEY)
}
export function getUser() {
  const raw = localStorage.getItem(USER_KEY)
  return raw ? JSON.parse(raw) : null
}
export function setUser(u) {
  if (u) localStorage.setItem(USER_KEY, JSON.stringify(u))
  else   localStorage.removeItem(USER_KEY)
}
export function clearAuth() {
  setToken('')
  setUser(null)
}

async function request(method, path, body, opts = {}) {
  const url = `${API_BASE}${path}`
  const headers = { 'Content-Type': 'application/json' }
  const token = getToken()
  if (token) headers['Authorization'] = `Bearer ${token}`
  if (opts.headers) Object.assign(headers, opts.headers)

  const init = { method, headers }
  if (body !== undefined) init.body = JSON.stringify(body)

  let res
  try {
    res = await fetch(url, init)
  } catch (e) {
    throw new Error('ไม่สามารถเชื่อมต่อ API Gateway ได้ (' + (e.message || e) + ')')
  }
  const text = await res.text()
  let data
  try { data = text ? JSON.parse(text) : {} } catch { data = { raw: text } }

  if (!res.ok) {
    if (res.status === 401) {
      clearAuth()
      window.dispatchEvent(new CustomEvent('sso:unauthorized'))
    }
    const body = data || {}
    const msg = body.error || res.statusText
    // แนบ body ทั้งก้อนไว้ใน error เพื่อให้ caller ใช้ hint/reason ได้
    const err = new Error(msg)
    err.backend = body
    err.status  = res.status
    throw err
  }
  return data
}

export const api = {
  // auth
  login:      (username, password) => request('POST', '/api/auth/login', { username, password }),
  logout:     ()                   => request('POST', '/api/auth/logout', {}),
  me:         ()                   => request('GET',  '/api/auth/me'),
  myEnvs:     ()                   => request('GET',  '/api/my-envs'),
  getEmployee:(username)           => request('GET',  `/api/employee/${encodeURIComponent(username)}`),

  // apps
  listApps:   ()                => request('GET',    '/api/apps'),
  createApp:  (body)            => request('POST',   '/api/apps', body),
  updateApp:  (id, body)        => request('PUT',    `/api/apps/${id}`, body),
  deleteApp:  (id)              => request('DELETE', `/api/apps/${id}`),

  // envs
  listEnvs:   (appId = 0)       => request('GET',    `/api/envs?appId=${appId}`),
  createEnv:  (body)            => request('POST',   '/api/envs', body),
  updateEnv:  (id, body)        => request('PUT',    `/api/envs/${id}`, body),
  deleteEnv:  (id)              => request('DELETE', `/api/envs/${id}`),

  // allowed ips
  listAllowedIPs:  (envId = 0)  => request('GET',    `/api/allowed-ips?envId=${envId}`),
  createAllowedIP: (body)        => request('POST',   '/api/allowed-ips', body),
  deleteAllowedIP: (id)          => request('DELETE', `/api/allowed-ips/${id}`),

  // allowed users
  listAllowedUsers:       (envId = 0) => request('GET',    `/api/allowed-users?envId=${envId}`),
  createAllowedUser:      (body)       => request('POST',   '/api/allowed-users', body),
  deleteAllowedUser:      (id)         => request('DELETE', `/api/allowed-users/${id}`),
  bulkCreateAllowedUsers: (body)       => request('POST',   '/api/allowed-users/bulk', body),
  listAllowedUsersByUser: (username)   => request('GET',    `/api/allowed-users/by-user?username=${encodeURIComponent(username)}`),

  // audit
  listAudit: (limit = 100)      => request('GET',    `/api/audit?limit=${limit}`),
}
