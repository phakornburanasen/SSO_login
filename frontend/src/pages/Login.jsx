import { useState } from 'react'
import { api, setToken, setUser } from '../api.js'

export default function Login({ onLogin }) {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading]   = useState(false)
  const [error, setError]       = useState('')

  const submit = async (e) => {
    e.preventDefault()
    setError('')
    if (!username.trim() || !password) {
      setError('กรุณากรอก username และ password')
      return
    }
    setLoading(true)
    try {
      const res = await api.login(username.trim(), password)
      setToken(res.token)
      const u = { username: res.username, displayName: res.displayName, expiresAt: res.expiresAt }
      setUser(u)
      onLogin(u)
    } catch (err) {
      setError(err.message || 'เข้าสู่ระบบไม่สำเร็จ')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-full flex items-center justify-center bg-gradient-to-br from-brand-700 via-brand-600 to-brand-500 px-4">
      <div className="w-full max-w-sm">
        <div className="text-center mb-6 text-white">
          <div className="inline-flex items-center justify-center w-14 h-14 rounded-full bg-white/15 backdrop-blur mb-3">
            <span className="text-2xl">🔐</span>
          </div>
          <h1 className="text-2xl font-bold">SSO Login</h1>
          <p className="text-sm text-white/80">Permission Controller</p>
        </div>

        <form onSubmit={submit} className="card p-6 space-y-4">
          <div>
            <label className="label" htmlFor="username">AD Username</label>
            <input
              id="username" type="text" autoComplete="username"
              className="input" placeholder="เช่น admin.sso"
              maxLength={15}
              value={username} onChange={e => setUsername(e.target.value)}
              autoFocus
            />
          </div>
          <div>
            <label className="label" htmlFor="password">Password</label>
            <input
              id="password" type="password" autoComplete="current-password"
              className="input" placeholder="••••••••"
              value={password} onChange={e => setPassword(e.target.value)}
            />
          </div>
          {error && (
            <div className="text-sm text-rose-600 bg-rose-50 border border-rose-200 rounded px-3 py-2">
              {error}
            </div>
          )}
          <button type="submit" disabled={loading} className="btn-primary w-full">
            {loading ? 'กำลังเข้าสู่ระบบ…' : 'เข้าสู่ระบบ'}
          </button>
          <p className="text-xs text-slate-500 text-center">
            เชื่อมต่อผ่าน API Gateway
          </p>
        </form>
      </div>
    </div>
  )
}
