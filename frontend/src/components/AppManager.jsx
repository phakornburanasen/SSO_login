import { useEffect, useState } from 'react'
import { api } from '../api.js'

const EMPTY = { code: '', name: '', description: '', active: true }

export default function AppManager() {
  const [items, setItems]   = useState([])
  const [form, setForm]     = useState(EMPTY)
  const [editing, setEditing] = useState(null)
  const [err, setErr]       = useState('')
  const [loading, setLoading] = useState(false)

  const load = async () => {
    setLoading(true); setErr('')
    try { const r = await api.listApps(); setItems(r.apps || []) }
    catch (e) { setErr(e.message) } finally { setLoading(false) }
  }
  useEffect(() => { load() }, [])

  const submit = async (e) => {
    e.preventDefault(); setErr('')
    if (!form.code.trim() || !form.name.trim()) return setErr('code/name required')
    try {
      if (editing === 'new') await api.createApp(form)
      else                   await api.updateApp(editing.id, form)
      setEditing(null); setForm(EMPTY); await load()
    } catch (e) { setErr(e.message) }
  }

  const remove = async (row) => {
    if (!window.confirm(`ลบ application "${row.code}" ?`)) return
    try { await api.deleteApp(row.id); await load() } catch (e) { setErr(e.message) }
  }

  return (
    <div>
      <div className="flex justify-end mb-3">
        <button className="btn-primary" onClick={() => { setEditing('new'); setForm(EMPTY); setErr('') }}>+ เพิ่ม Application</button>
      </div>
      {err && <div className="mb-3 text-sm text-rose-600 bg-rose-50 border border-rose-200 rounded px-3 py-2">{err}</div>}
      <div className="card overflow-x-auto">
        <table className="table">
          <thead>
            <tr>
              <th className="th">Code</th>
              <th className="th">Name</th>
              <th className="th">Description</th>
              <th className="th">Active</th>
              <th className="th w-32">จัดการ</th>
            </tr>
          </thead>
          <tbody>
            {loading && <tr><td colSpan={5} className="text-center py-4 text-slate-500">กำลังโหลด…</td></tr>}
            {!loading && items.length === 0 && <tr><td colSpan={5} className="text-center py-4 text-slate-500">ยังไม่มีข้อมูล</td></tr>}
            {items.map(row => (
              <tr key={row.id} className="hover:bg-slate-50">
                <td className="td font-mono">{row.code}</td>
                <td className="td">{row.name}</td>
                <td className="td text-slate-600">{row.description}</td>
                <td className="td">{row.active ? <span className="badge-on">Active</span> : <span className="badge-off">Inactive</span>}</td>
                <td className="td">
                  <div className="flex gap-2">
                    <button className="btn-ghost px-2 py-1 text-xs" onClick={() => { setEditing(row); setForm({ code: row.code, name: row.name, description: row.description, active: row.active }) }}>แก้ไข</button>
                    <button className="btn-ghost px-2 py-1 text-xs text-rose-600" onClick={() => remove(row)}>ลบ</button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {editing !== null && (
        <form onSubmit={submit} className="card p-5 mt-4 space-y-3">
          <h3 className="font-semibold">{editing === 'new' ? 'เพิ่ม Application' : `แก้ไข Application #${editing.id}`}</h3>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="label">Code *</label>
              <input className="input font-mono" value={form.code} disabled={editing !== 'new'}
                     onChange={e => setForm({ ...form, code: e.target.value.toUpperCase() })}/>
            </div>
            <div>
              <label className="label">Name *</label>
              <input className="input" value={form.name} onChange={e => setForm({ ...form, name: e.target.value })}/>
            </div>
            <div className="col-span-2">
              <label className="label">Description</label>
              <input className="input" value={form.description} onChange={e => setForm({ ...form, description: e.target.value })}/>
            </div>
            <label className="inline-flex items-center gap-2 text-sm">
              <input type="checkbox" checked={form.active} onChange={e => setForm({ ...form, active: e.target.checked })}/>
              <span>Active</span>
            </label>
          </div>
          <div className="flex justify-end gap-2">
            <button type="button" className="btn-secondary" onClick={() => { setEditing(null); setErr('') }}>ยกเลิก</button>
            <button type="submit" className="btn-primary">บันทึก</button>
          </div>
        </form>
      )}
    </div>
  )
}
