import { useEffect, useState } from 'react'
import { api } from '../api.js'

const EMPTY = { envId: 0, ipCidr: '', description: '' }

export default function AllowedIpManager() {
  const [envs, setEnvs]   = useState([])
  const [items, setItems] = useState([])
  const [filterEnv, setFilterEnv] = useState(0)
  const [form, setForm]   = useState(EMPTY)
  const [err, setErr]     = useState('')
  const [loading, setLoading] = useState(false)

  const load = async () => {
    setLoading(true); setErr('')
    try {
      const [e, i] = await Promise.all([api.listEnvs(0), api.listAllowedIPs(filterEnv)])
      setEnvs(e.envs || [])
      setItems(i.allowedIps || [])
    } catch (e) { setErr(e.message) } finally { setLoading(false) }
  }
  useEffect(() => { load() }, [filterEnv])

  const submit = async (e) => {
    e.preventDefault(); setErr('')
    if (!form.envId || !form.ipCidr.trim()) return setErr('envId and ipCidr required')
    try {
      await api.createAllowedIP({ ...form, createdBy: 'admin-ui' })
      setForm({ ...EMPTY, envId: form.envId })
      await load()
    } catch (e) { setErr(e.message) }
  }

  const remove = async (row) => {
    if (!window.confirm(`ลบ IP "${row.ipCidr}" ?`)) return
    try { await api.deleteAllowedIP(row.id); await load() } catch (e) { setErr(e.message) }
  }

  return (
    <div>
      <div className="flex flex-wrap items-end gap-3 mb-4">
        <div>
          <label className="label">กรองตาม Environment</label>
          <select className="input min-w-[260px]" value={filterEnv} onChange={e => setFilterEnv(Number(e.target.value))}>
            <option value={0}>— ทั้งหมด —</option>
            {envs.map(en => <option key={en.id} value={en.id}>{en.appCode}/{en.envCode} — {en.envName}</option>)}
          </select>
        </div>
      </div>

      <form onSubmit={submit} className="card p-4 mb-4 grid grid-cols-1 md:grid-cols-4 gap-3">
        <div className="md:col-span-1">
          <label className="label">Environment *</label>
          <select className="input" value={form.envId} onChange={e => setForm({ ...form, envId: Number(e.target.value) })}>
            <option value={0}>— เลือก —</option>
            {envs.map(en => <option key={en.id} value={en.id}>{en.appCode}/{en.envCode}</option>)}
          </select>
        </div>
        <div className="md:col-span-1">
          <label className="label">IP / CIDR *</label>
          <input className="input font-mono" value={form.ipCidr} placeholder="10.0.32.0/24"
                 onChange={e => setForm({ ...form, ipCidr: e.target.value })}/>
        </div>
        <div className="md:col-span-1">
          <label className="label">Description</label>
          <input className="input" value={form.description} onChange={e => setForm({ ...form, description: e.target.value })}/>
        </div>
        <div className="md:col-span-1 flex items-end">
          <button className="btn-primary w-full" type="submit">+ เพิ่ม</button>
        </div>
      </form>

      {err && <div className="mb-3 text-sm text-rose-600 bg-rose-50 border border-rose-200 rounded px-3 py-2">{err}</div>}

      <div className="card overflow-x-auto">
        <table className="table">
          <thead>
            <tr>
              <th className="th">Env</th>
              <th className="th">IP / CIDR</th>
              <th className="th">Description</th>
              <th className="th">Active</th>
              <th className="th">Created</th>
              <th className="th w-20">จัดการ</th>
            </tr>
          </thead>
          <tbody>
            {loading && <tr><td colSpan={6} className="text-center py-4 text-slate-500">กำลังโหลด…</td></tr>}
            {!loading && items.length === 0 && <tr><td colSpan={6} className="text-center py-4 text-slate-500">ยังไม่มีข้อมูล</td></tr>}
            {items.map(row => {
              const en = envs.find(e => e.id === row.envId)
              return (
                <tr key={row.id} className="hover:bg-slate-50">
                  <td className="td font-mono text-xs">{en ? `${en.appCode}/${en.envCode}` : `#${row.envId}`}</td>
                  <td className="td font-mono">{row.ipCidr}</td>
                  <td className="td text-slate-600">{row.description}</td>
                  <td className="td">{row.active ? <span className="badge-on">Active</span> : <span className="badge-off">Off</span>}</td>
                  <td className="td text-xs text-slate-500">{row.createdBy || '-'}</td>
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
