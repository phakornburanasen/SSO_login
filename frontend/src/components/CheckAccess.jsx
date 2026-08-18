import { useState } from 'react'
import { API_BASE } from '../api.js'

const SAMPLE = 'http://10.0.32.71/HelpDesk/'

export default function CheckAccess() {
  const [form, setForm] = useState({ baseUrl: SAMPLE, clientIp: '', adUsername: '' })
  const [result, setResult] = useState(null)
  const [err, setErr] = useState('')
  const [loading, setLoading] = useState(false)

  const submit = async (e) => {
    e.preventDefault()
    setErr('')
    setResult(null)
    setLoading(true)
    try {
      const r = await fetch(`${API_BASE}/api/check-access`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          baseUrl: form.baseUrl,
          clientIp: form.clientIp,
          adUsername: form.adUsername,
        }),
      })
      const data = await r.json()
      setResult({ status: r.status, data })
    } catch (e) {
      setErr(e.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="grid gap-6 xl:grid-cols-[minmax(0,0.9fr)_minmax(320px,0.7fr)]">
      <div className="card p-6">
        <div className="mb-5">
          <h2 className="section-title">จำลองการตรวจสิทธิ์</h2>
          <p className="section-desc">
            ทดสอบเรียก <code className="rounded-lg bg-slate-100 px-1.5 py-0.5 font-mono text-xs">POST /api/check-access</code> เพื่อดูผลลัพธ์จริงตาม
            base URL, client IP และ AD username
          </p>
        </div>

        <form onSubmit={submit} className="space-y-4">
          <div>
            <label className="label">Base URL</label>
            <input className="input font-mono" value={form.baseUrl} onChange={(e) => setForm({ ...form, baseUrl: e.target.value })} />
          </div>
          <div>
            <label className="label">Client IP</label>
            <input className="input font-mono" value={form.clientIp} onChange={(e) => setForm({ ...form, clientIp: e.target.value })} placeholder="10.0.32.100" />
          </div>
          <div>
            <label className="label">AD Username</label>
            <input className="input font-mono" value={form.adUsername} onChange={(e) => setForm({ ...form, adUsername: e.target.value })} placeholder="somchai.s" />
          </div>
          <button className="btn-primary" disabled={loading}>
            {loading ? (
              <span className="flex items-center gap-2">
                <Spinner /> กำลังตรวจสิทธิ์…
              </span>
            ) : (
              'ตรวจสิทธิ์'
            )}
          </button>
        </form>

        {err && <div className="error-alert mt-4 fade-in">{err}</div>}
      </div>

      <div className="card p-6">
        <div className="mb-4">
          <h3 className="section-title">ผลลัพธ์จาก API</h3>
          <p className="section-desc">ดู status code และ payload ที่ backend ส่งกลับมาเพื่อตรวจสอบการตั้งค่าได้ทันที</p>
        </div>

        {result ? (
          <div className="space-y-4 fade-in">
            <div className="inline-flex items-center gap-2 rounded-xl bg-slate-100 px-3 py-1.5 text-sm font-medium text-slate-600">
              HTTP {result.status}
            </div>
            <pre className="overflow-x-auto rounded-xl bg-slate-900 p-4 text-xs leading-6 text-slate-100">
{JSON.stringify(result.data, null, 2)}
            </pre>
          </div>
        ) : (
          <div className="rounded-xl border-2 border-dashed border-slate-200 bg-slate-50/50 px-5 py-10 text-center text-sm text-slate-400">
            เมื่อกดตรวจสิทธิ์แล้ว ผลลัพธ์จาก API จะมาแสดงตรงนี้
          </div>
        )}
      </div>
    </div>
  )
}

function Spinner() {
  return (
    <svg className="h-4 w-4 animate-spin" viewBox="0 0 24 24" fill="none">
      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
    </svg>
  )
}
