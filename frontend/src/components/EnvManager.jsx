import { useEffect, useState } from 'react'
import { api } from '../api.js'

const EMPTY = {
  appId: 0,
  envCode: '',
  envName: '',
  baseUrl: '',
  hostIp: '',
  basePath: '',
  adUser: '',
  active: true,
}

export default function EnvManager() {
  const [apps, setApps]         = useState([])
  const [envs, setEnvs]         = useState([])
  const [filterApp, setFilterApp] = useState(0)
  const [editing, setEditing]   = useState(null)  // null | 'new' | env object
  const [form, setForm]         = useState(EMPTY)
  const [err, setErr]           = useState('')
  const [loading, setLoading]   = useState(false)
  const [saving, setSaving]     = useState(false)

  const load = async () => {
    setLoading(true); setErr('')
    try {
      const [a, e] = await Promise.all([api.listApps(), api.listEnvs(filterApp)])
      setApps(a.apps || [])
      setEnvs(e.envs || [])
    } catch (e) { setErr(e.message) }
    finally { setLoading(false) }
  }

  useEffect(() => { load() }, [filterApp])

  const openNew = () => { setEditing('new'); setForm({ ...EMPTY, appId: apps[0]?.id || 0 }); setErr('') }
  const openEdit = (row) => {
    setEditing(row)
    setForm({
      appId:   row.appId,
      envCode: row.envCode,
      envName: row.envName,
      baseUrl: row.baseUrl,
      hostIp:  row.hostIp || '',
      basePath: row.basePath || '',
      adUser:  row.adUser || '',
      active:  row.active,
    })
    setErr('')
  }
  const cancel = () => { setEditing(null); setErr('') }

  const submit = async (e) => {
    e.preventDefault()
    setErr('')
    if (!form.appId)               return setErr('กรุณาเลือก Application')
    if (!form.envCode.trim())      return setErr('กรุณากรอก Env Code')
    if (!form.envName.trim())      return setErr('กรุณากรอก Env Name')
    if (!form.baseUrl.trim())      return setErr('กรุณากรอก Base URL')

    const body = {
      appId:   Number(form.appId),
      envCode: form.envCode.trim().toUpperCase(),
      envName: form.envName.trim(),
      baseUrl: form.baseUrl.trim(),
      hostIp:  form.hostIp.trim(),
      basePath: form.basePath.trim(),
      adUser:  form.adUser.trim(),
      active:  !!form.active,
    }
    setSaving(true)
    try {
      if (editing === 'new') await api.createEnv(body)
      else                   await api.updateEnv(editing.id, body)
      setEditing(null)
      await load()
    } catch (e) { setErr(e.message) }
    finally { setSaving(false) }
  }

  const remove = async (row) => {
    if (!window.confirm(`ลบ env "${row.envName}" ?`)) return
    setErr('')
    try { await api.deleteEnv(row.id); await load() }
    catch (e) { setErr(e.message) }
  }

  return (
    <div>
      <div className="flex flex-wrap items-center gap-3 mb-4">
        <div>
          <label className="label">กรองตาม Application</label>
          <select className="input min-w-[200px]"
                  value={filterApp} onChange={e => setFilterApp(Number(e.target.value))}>
            <option value={0}>— ทั้งหมด —</option>
            {apps.map(a => <option key={a.id} value={a.id}>{a.appCode} — {a.name}</option>)}
          </select>
        </div>
        <div className="ml-auto self-end">
          <button className="btn-primary" onClick={openNew} disabled={apps.length === 0}>
            + เพิ่ม Environment
          </button>
        </div>
      </div>

      {err && <div className="mb-3 text-sm text-rose-600 bg-rose-50 border border-rose-200 rounded px-3 py-2">{err}</div>}

      <div className="card overflow-x-auto">
        <table className="table">
          <thead>
            <tr>
              <th className="th">App</th>
              <th className="th">Env</th>
              <th className="th">Name</th>
              <th className="th">Base URL</th>
              <th className="th">Host IP</th>
              <th className="th">Base Path</th>
              <th className="th">AD User</th>
              <th className="th">Active</th>
              <th className="th w-32">จัดการ</th>
            </tr>
          </thead>
          <tbody>
            {loading && <tr><td colSpan={9} className="text-center text-slate-500 py-4">กำลังโหลด…</td></tr>}
            {!loading && envs.length === 0 && <tr><td colSpan={9} className="text-center text-slate-500 py-4">ยังไม่มีข้อมูล</td></tr>}
            {envs.map(row => (
              <tr key={row.id} className="hover:bg-slate-50">
                <td className="td font-mono text-xs">{row.appCode}</td>
                <td className="td font-mono">{row.envCode}</td>
                <td className="td">{row.envName}</td>
                <td className="td font-mono text-xs break-all">{row.baseUrl}</td>
                <td className="td font-mono">{row.hostIp || '-'}</td>
                <td className="td font-mono">{row.basePath || '-'}</td>
                <td className="td font-mono">{row.adUser || '-'}</td>
                <td className="td">{row.active ? <span className="badge-on">Active</span> : <span className="badge-off">Inactive</span>}</td>
                <td className="td">
                  <div className="flex gap-2">
                    <button className="btn-ghost px-2 py-1 text-xs" onClick={() => openEdit(row)}>แก้ไข</button>
                    <button className="btn-ghost px-2 py-1 text-xs text-rose-600" onClick={() => remove(row)}>ลบ</button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {editing !== null && (
        <Modal title={editing === 'new' ? 'เพิ่ม Environment' : `แก้ไข Environment #${editing.id}`} onClose={cancel}>
          <form onSubmit={submit} className="space-y-3">
            {err && <div className="text-sm text-rose-600 bg-rose-50 border border-rose-200 rounded px-3 py-2">{err}</div>}
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="label">Application *</label>
                <select className="input" value={form.appId}
                        onChange={e => setForm({ ...form, appId: Number(e.target.value) })} disabled={editing !== 'new'}>
                  <option value={0}>— เลือก —</option>
                  {apps.map(a => <option key={a.id} value={a.id}>{a.appCode} — {a.name}</option>)}
                </select>
              </div>
              <div>
                <label className="label">Env Code *</label>
                <input className="input" value={form.envCode} maxLength={20}
                       onChange={e => setForm({ ...form, envCode: e.target.value.toUpperCase() })}/>
              </div>
              <div className="col-span-2">
                <label className="label">Env Name *</label>
                <input className="input" value={form.envName}
                       onChange={e => setForm({ ...form, envName: e.target.value })}/>
              </div>
              <div className="col-span-2">
                <label className="label">Base URL * <span className="text-slate-400 font-normal">(http://ip/base/)</span></label>
                <input className="input font-mono" value={form.baseUrl}
                       placeholder="http://10.0.32.71/HelpDesk/"
                       onChange={e => setForm({ ...form, baseUrl: e.target.value })}/>
              </div>
              <div>
                <label className="label">Host IP</label>
                <input className="input font-mono" value={form.hostIp}
                       placeholder="10.0.32.71"
                       onChange={e => setForm({ ...form, hostIp: e.target.value })}/>
              </div>
              <div>
                <label className="label">Base Path</label>
                <input className="input font-mono" value={form.basePath}
                       placeholder="/HelpDesk/"
                       onChange={e => setForm({ ...form, basePath: e.target.value })}/>
              </div>
              <div>
                <label className="label">AD User <span className="text-slate-400 font-normal">(≤15)</span></label>
                <input className="input font-mono" value={form.adUser} maxLength={15}
                       onChange={e => setForm({ ...form, adUser: e.target.value })}/>
              </div>
              <div className="flex items-end">
                <label className="inline-flex items-center gap-2 text-sm">
                  <input type="checkbox" checked={form.active}
                         onChange={e => setForm({ ...form, active: e.target.checked })}/>
                  <span>Active</span>
                </label>
              </div>
            </div>
            <div className="flex justify-end gap-2 pt-2 border-t border-slate-100">
              <button type="button" className="btn-secondary" onClick={cancel}>ยกเลิก</button>
              <button type="submit" className="btn-primary" disabled={saving}>
                {saving ? 'กำลังบันทึก…' : 'บันทึก'}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}

function Modal({ title, onClose, children }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4" onClick={onClose}>
      <div className="bg-white rounded-lg shadow-xl w-full max-w-xl" onClick={e => e.stopPropagation()}>
        <div className="px-5 py-3 border-b border-slate-200 flex items-center justify-between">
          <h2 className="font-semibold">{title}</h2>
          <button onClick={onClose} className="text-slate-500 hover:text-slate-800 text-xl leading-none">×</button>
        </div>
        <div className="p-5">{children}</div>
      </div>
    </div>
  )
}
