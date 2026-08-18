import { useEffect, useState } from 'react'
import { api } from '../api.js'

export default function AuditLog() {
  const [items, setItems] = useState([])
  const [err, setErr]     = useState('')
  const [loading, setLoading] = useState(false)

  const load = async () => {
    setLoading(true); setErr('')
    try { const r = await api.listAudit(200); setItems(r.audit || []) }
    catch (e) { setErr(e.message) } finally { setLoading(false) }
  }
  useEffect(() => { load() }, [])

  const color = (r) => {
    if (r === 'ALLOW') return 'bg-emerald-100 text-emerald-700'
    if (r && r.startsWith('DENY')) return 'bg-rose-100 text-rose-700'
    return 'bg-amber-100 text-amber-700'
  }

  return (
    <div>
      <div className="flex justify-end mb-3">
        <button className="btn-secondary" onClick={load} disabled={loading}>↻ รีเฟรช</button>
      </div>
      {err && <div className="mb-3 text-sm text-rose-600 bg-rose-50 border border-rose-200 rounded px-3 py-2">{err}</div>}
      <div className="card overflow-x-auto">
        <table className="table">
          <thead>
            <tr>
              <th className="th">Time</th>
              <th className="th">App/Env</th>
              <th className="th">Base URL</th>
              <th className="th">Client IP</th>
              <th className="th">AD User</th>
              <th className="th">Result</th>
              <th className="th">Reason</th>
            </tr>
          </thead>
          <tbody>
            {loading && <tr><td colSpan={7} className="text-center py-4 text-slate-500">กำลังโหลด…</td></tr>}
            {!loading && items.length === 0 && <tr><td colSpan={7} className="text-center py-4 text-slate-500">ยังไม่มีข้อมูล</td></tr>}
            {items.map(r => (
              <tr key={r.id} className="hover:bg-slate-50">
                <td className="td text-xs text-slate-500">{new Date(r.createdAt).toLocaleString('th-TH')}</td>
                <td className="td font-mono text-xs">{r.appCode}/{r.envCode}</td>
                <td className="td font-mono text-xs break-all max-w-xs">{r.baseUrl}</td>
                <td className="td font-mono">{r.clientIp}</td>
                <td className="td font-mono">{r.adUsername}</td>
                <td className="td"><span className={'badge ' + color(r.result)}>{r.result}</span></td>
                <td className="td text-xs text-slate-500">{r.denyReason}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
