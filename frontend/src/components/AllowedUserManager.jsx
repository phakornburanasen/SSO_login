import { useEffect, useState } from 'react'
import { api } from '../api.js'

const EMPTY = {
  envId: 0,
  adUsername: '',
  employeeId: '',
  displayName: '',
  email: '',
  department: '',
  grantedBy: 'admin-ui',
  expiresAt: '',
}

export default function AllowedUserManager() {
  const [envs, setEnvs]   = useState([])
  const [items, setItems] = useState([])
  const [filterEnv, setFilterEnv] = useState(0)
  const [form, setForm]   = useState(EMPTY)
  const [err, setErr]     = useState('')
  const [loading, setLoading] = useState(false)

  const load = async () => {
    setLoading(true); setErr('')
    try {
      const [e, u] = await Promise.all([api.listEnvs(0), api.listAllowedUsers(filterEnv)])
      setEnvs(e.envs || [])
      setItems(u.allowedUsers || [])
    } catch (e) { setErr(e.message) } finally { setLoading(false) }
  }
  useEffect(() => { load() }, [filterEnv])

  const submit = async (e) => {
    e.preventDefault(); setErr('')
    if (!form.envId || !form.adUsername.trim()) return setErr('envId and adUsername required')
    const body = { ...form }
    if (!body.expiresAt) delete body.expiresAt
    try {
      await api.createAllowedUser(body)
      setForm({ ...EMPTY, envId: form.envId })
      await load()
    } catch (e) { setErr(e.message) }
  }

  const remove = async (row) => {
    if (!window.confirm(`ลบ AD user "${row.adUsername}" ?`)) return
    try { await api.deleteAllowedUser(row.id); await load() } catch (e) { setErr(e.message) }
  }

  return (
    <div>
      <div className="mb-4">
        <label className="label">กรองตาม Environment</label>
        <select className="input min-w-[260px]" value={filterEnv} onChange={e => setFilterEnv(Number(e.target.value))}>
          <option value={0}>— ทั้งหมด —</option>
          {envs.map(en => <option key={en.id} value={en.id}>{en.appCode}/{en.envCode} — {en.envName}</option>)}
        </select>
      </div>

      <form onSubmit={submit} className="card p-4 mb-4 grid grid-cols-1 md:grid-cols-4 gap-3">
        <div>
          <label className="label">Environment *</label>
          <select className="input" value={form.envId} onChange={e => setForm({ ...form, envId: Number(e.target.value) })}>
            <option value={0}>— เลือก —</option>
            {envs.map(en => <option key={en.id} value={en.id}>{en.appCode}/{en.envCode}</option>)}
          </select>
        </div>
        <div>
          <label className="label">AD Username *</label>
          <input className="input font-mono" value={form.adUsername} onChange={e => setForm({ ...form, adUsername: e.target.value })}/>
        </div>
        <div>
          <label className="label">Employee ID</label>
          <input className="input font-mono" value={form.employeeId} onChange={e => setForm({ ...form, employeeId: e.target.value })}/>
        </div>
        <div>
          <label className="label">Display Name</label>
          <input className="input" value={form.displayName} onChange={e => setForm({ ...form, displayName: e.target.value })}/>
        </div>
        <div>
          <label className="label">Email</label>
          <input className="input" value={form.email} onChange={e => setForm({ ...form, email: e.target.value })}/>
        </div>
        <div>
          <label className="label">Department</label>
          <input className="input" value={form.department} onChange={e => setForm({ ...form, department: e.target.value })}/>
        </div>
        <div>
          <label className="label">Expires At</label>
          <input type="datetime-local" className="input" value={form.expiresAt} onChange={e => setForm({ ...form, expiresAt: e.target.value })}/>
        </div>
        <div className="flex items-end">
          <button className="btn-primary w-full" type="submit">+ เพิ่ม AD user</button>
        </div>
      </form>

      {err && <div className="mb-3 text-sm text-rose-600 bg-rose-50 border border-rose-200 rounded px-3 py-2">{err}</div>}

      <div className="card overflow-x-auto">
        <table className="table">
          <thead>
            <tr>
              <th className="th">Env</th>
              <th className="th">AD Username</th>
              <th className="th">Employee</th>
              <th className="th">Display Name</th>
              <th className="th">Email</th>
              <th className="th">Dept</th>
              <th className="th">Expires</th>
              <th className="th w-20">จัดการ</th>
            </tr>
          </thead>
          <tbody>
            {loading && <tr><td colSpan={8} className="text-center py-4 text-slate-500">กำลังโหลด…</td></tr>}
            {!loading && items.length === 0 && <tr><td colSpan={8} className="text-center py-4 text-slate-500">ยังไม่มีข้อมูล</td></tr>}
            {items.map(row => {
              const en = envs.find(e => e.id === row.envId)
              return (
                <tr key={row.id} className="hover:bg-slate-50">
                  <td className="td font-mono text-xs">{en ? `${en.appCode}/${en.envCode}` : `#${row.envId}`}</td>
                  <td className="td font-mono">{row.adUsername}</td>
                  <td className="td font-mono">{row.employeeId || '-'}</td>
                  <td className="td">{row.displayName || '-'}</td>
                  <td className="td text-slate-600 text-xs">{row.email || '-'}</td>
                  <td className="td">{row.department || '-'}</td>
                  <td className="td text-xs">{row.expiresAt ? new Date(row.expiresAt).toLocaleString('th-TH') : '-'}</td>
                  <td className="td">
                    <button className="btn-ghost px-2 py-1 text-xs text-rose-600" onClick={() => remove(row)}>ลบ</button>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </div>
  )
}
