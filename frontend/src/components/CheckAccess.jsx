import { useState } from 'react'
import { api } from '../api.js'

const SAMPLE = 'http://10.0.32.71/HelpDesk/'

export default function CheckAccess() {
  const [form, setForm] = useState({ baseUrl: SAMPLE, clientIp: '', adUsername: '' })
  const [result, setResult] = useState(null)
  const [err, setErr] = useState('')
  const [loading, setLoading] = useState(false)

  const submit = async (e) => {
    e.preventDefault(); setErr(''); setResult(null)
    setLoading(true)
    try {
      const r = await fetch(`${(import.meta.env.VITE_API_BASE || 'http://127.0.0.1:18000/api/SSO_login')}/api/check-access`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          baseUrl:    form.baseUrl,
          clientIp:   form.clientIp,
          adUsername: form.adUsername,
        }),
      })
      const data = await r.json()
      setResult({ status: r.status, data })
    } catch (e) { setErr(e.message) }
    finally { setLoading(false) }
  }

  return (
    <div className="max-w-2xl">
      <p className="text-sm text-slate-600 mb-4">
        ทดสอบเรียก <code className="bg-slate-100 px-1 rounded">POST /api/check-access</code> เพื่อตรวจสิทธิ์ตาม <b>baseUrl + clientIp + adUsername</b>
      </p>
      <form onSubmit={submit} className="card p-5 space-y-3">
        <div>
          <label className="label">Base URL</label>
          <input className="input font-mono" value={form.baseUrl} onChange={e => setForm({ ...form, baseUrl: e.target.value })}/>
        </div>
        <div>
          <label className="label">Client IP</label>
          <input className="input font-mono" value={form.clientIp} onChange={e => setForm({ ...form, clientIp: e.target.value })} placeholder="10.0.32.100"/>
        </div>
        <div>
          <label className="label">AD Username</label>
          <input className="input font-mono" value={form.adUsername} onChange={e => setForm({ ...form, adUsername: e.target.value })} placeholder="somchai.s"/>
        </div>
        <button className="btn-primary" disabled={loading}>{loading ? 'กำลังตรวจ…' : 'ตรวจสิทธิ์'}</button>
      </form>

      {err && <div className="mt-3 text-sm text-rose-600 bg-rose-50 border border-rose-200 rounded px-3 py-2">{err}</div>}

      {result && (
        <div className="mt-4 card p-4">
          <div className="text-sm text-slate-500 mb-1">HTTP {result.status}</div>
          <pre className="text-xs font-mono bg-slate-900 text-slate-100 rounded p-3 overflow-x-auto">
{JSON.stringify(result.data, null, 2)}
          </pre>
        </div>
      )}
    </div>
  )
}
