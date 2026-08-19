import { useEffect, useMemo, useState } from 'react'
import { api, getUser } from '../api.js'
import Pagination from './Pagination.jsx'

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
  const currentUser = getUser()
  const isAdmin = currentUser?.role === 'admin'
  const accessibleEnvIds = useMemo(
    () => new Set((currentUser?.accessibleEnvs || []).map(e => e.id)),
    [currentUser]
  )

  const [apps, setApps] = useState([])
  const [envs, setEnvs] = useState([])
  const [filterApp, setFilterApp] = useState(0)
  const [editing, setEditing] = useState(null)
  const [form, setForm] = useState(EMPTY)
  const [err, setErr] = useState('')
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [deleteTarget, setDeleteTarget] = useState(null)

  const pagedEnvs = useMemo(() => {
    const start = (page - 1) * pageSize
    return envs.slice(start, start + pageSize)
  }, [envs, page, pageSize])

  const load = async () => {
    setLoading(true)
    setErr('')
    try {
      const [a, e] = await Promise.all([api.listApps(), api.listEnvs(filterApp)])
      const allEnvs = e.envs || []
      // กรอง: ถ้าไม่ใช่ admin → เห็นเฉพาะ env ที่อยู่ใน accessibleEnvs
      const filtered = isAdmin
        ? allEnvs
        : allEnvs.filter((env) => accessibleEnvIds.has(env.id))
      setApps(a.apps || [])
      setEnvs(filtered)
    } catch (e) {
      setErr(e.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [filterApp])

  useEffect(() => {
    const totalPages = Math.max(1, Math.ceil(envs.length / pageSize))
    if (page > totalPages) setPage(totalPages)
  }, [envs.length, page, pageSize])

  const openNew = () => {
    setEditing(null)
    // ถ้าเป็น user ทั่วไป → pre-fill adUser ด้วย username ตัวเอง
    const defaultAdUser = isAdmin ? '' : (currentUser?.username || '')
    setForm({ ...EMPTY, appId: apps[0]?.id || 0, adUser: defaultAdUser })
    setErr('')
    setModalOpen(true)
  }

  const openEdit = (row) => {
    setEditing(row)
    setForm({
      appId: row.appId,
      envCode: row.envCode,
      envName: row.envName,
      baseUrl: row.baseUrl,
      hostIp: row.hostIp || '',
      basePath: row.basePath || '',
      adUser: row.adUser || '',
      active: row.active,
    })
    setErr('')
    setModalOpen(true)
  }

  const closeModal = () => {
    setModalOpen(false)
    setEditing(null)
    setForm(EMPTY)
    setErr('')
  }

  const submit = async (e) => {
    e.preventDefault()
    setErr('')
    if (!form.appId) return setErr('กรุณาเลือก Application')
    if (!form.envCode.trim()) return setErr('กรุณากรอก Env Code')
    if (!form.envName.trim()) return setErr('กรุณากรอก Env Name')
    if (!form.baseUrl.trim()) return setErr('กรุณากรอก Base URL')

    const body = {
      appId: Number(form.appId),
      envCode: form.envCode.trim().toUpperCase(),
      envName: form.envName.trim(),
      baseUrl: form.baseUrl.trim(),
      hostIp: form.hostIp.trim(),
      basePath: form.basePath.trim(),
      adUser: form.adUser.trim(),
      active: !!form.active,
    }

    setSaving(true)
    try {
      if (editing) await api.updateEnv(editing.id, body)
      else await api.createEnv(body)
      closeModal()
      await load()
    } catch (e) {
      setErr(e.message)
    } finally {
      setSaving(false)
    }
  }

  const confirmDelete = (row) => {
    setDeleteTarget(row)
  }

  const doDelete = async () => {
    if (!deleteTarget) return
    try {
      await api.deleteEnv(deleteTarget.id)
      setDeleteTarget(null)
      await load()
    } catch (e) {
      setErr(e.message)
    }
  }

  return (
    <div className="space-y-6">
      <div className="table-toolbar">
        <div>
          <h2 className="section-title">
            {isAdmin ? 'Environment ทั้งหมด' : 'Environment ที่คุณดูแล'}
          </h2>
          <p className="section-desc">
            {isAdmin
              ? 'จัดการ base URL, host IP และ AD user สำหรับแต่ละ environment ของระบบ'
              : 'คุณเห็นเฉพาะ environment ที่คุณเป็น ADUser เท่านั้น — สิทธิ์ถูกควบคุมโดยฐานข้อมูล sso_environments.ADUser'}
          </p>
        </div>

        <div className="subcard ml-auto flex flex-wrap items-end gap-3 px-4 py-4">
          <div>
            <label className="label">กรองตาม Application</label>
            <select
              className="input min-w-[240px]"
              value={filterApp}
              onChange={(e) => {
                setFilterApp(Number(e.target.value))
                setPage(1)
              }}
            >
              <option value={0}>ทั้งหมด</option>
              {apps.map((app) => (
                <option key={app.id} value={app.id}>
                  {app.code} — {app.name}
                </option>
              ))}
            </select>
          </div>
          <button type="button" className="btn-primary" onClick={openNew} disabled={apps.length === 0}>
            + เพิ่ม Environment
          </button>
        </div>
      </div>

      {!isAdmin && (
        <div className="info-alert">
          <svg className="h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span>
            คุณดูแลได้ <strong>{currentUser?.accessibleEnvs?.length || 0}</strong> environment
            — ระบบ login เช็คสิทธิ์จาก <code className="font-mono text-xs">sso_environments.ADUser</code> ในฐานข้อมูล
          </span>
        </div>
      )}

      {err && <div className="error-alert fade-in">{err}</div>}

      <div className="table-shell">
        <div className="overflow-x-auto">
          <table className="table">
            <thead>
              <tr>
                <th className="th">Application</th>
                <th className="th">Env</th>
                <th className="th">Name</th>
                <th className="th">Base URL</th>
                <th className="th">Host IP</th>
                <th className="th">Base Path</th>
                <th className="th">AD User</th>
                <th className="th">Status</th>
                <th className="th w-[180px]">จัดการ</th>
              </tr>
            </thead>
            <tbody>
              {loading && (
                <tr>
                  <td colSpan={9} className="td py-10 text-center text-slate-400">
                    <div className="flex items-center justify-center gap-2">
                      <Spinner /> กำลังโหลดข้อมูล…
                    </div>
                  </td>
                </tr>
              )}
              {!loading && envs.length === 0 && (
                <tr>
                  <td colSpan={9} className="td py-10 text-center text-slate-400">ยังไม่มี Environment</td>
                </tr>
              )}
              {!loading &&
                pagedEnvs.map((row) => (
                  <tr key={row.id} className="table-row-hover fade-in">
                    <td className="td">
                      <span className="inline-flex rounded-lg bg-slate-100 px-2.5 py-1 text-xs font-semibold text-slate-700">
                        {row.appCode}
                      </span>
                    </td>
                    <td className="td">
                      <span className="font-mono font-semibold text-slate-900">{row.envCode}</span>
                    </td>
                    <td className="td font-medium text-slate-900">{row.envName}</td>
                    <td className="td max-w-[320px] break-all font-mono text-xs text-slate-600">{row.baseUrl}</td>
                    <td className="td font-mono text-xs">{row.hostIp || '-'}</td>
                    <td className="td font-mono text-xs">{row.basePath || '-'}</td>
                    <td className="td font-mono text-xs">{row.adUser || '-'}</td>
                    <td className="td">
                      {row.active ? (
                        <span className="badge-on">Active</span>
                      ) : (
                        <span className="badge-off">Inactive</span>
                      )}
                    </td>
                    <td className="td">
                      <div className="flex gap-2">
                        <button type="button" className="btn-secondary px-3 py-1.5 text-xs" onClick={() => openEdit(row)}>
                          แก้ไข
                        </button>
                        <button
                          type="button"
                          className="btn-ghost px-3 py-1.5 text-xs text-rose-600 hover:text-rose-700"
                          onClick={() => confirmDelete(row)}
                        >
                          ลบ
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
            </tbody>
          </table>
        </div>

        <Pagination
          page={page}
          totalItems={envs.length}
          pageSize={pageSize}
          onPageChange={setPage}
          onPageSizeChange={(size) => {
            setPageSize(size)
            setPage(1)
          }}
        />
      </div>

      {/* Add/Edit Modal */}
      {modalOpen && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal-content modal-animate" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <div>
                <h2 className="text-lg font-semibold text-slate-900">
                  {editing ? `แก้ไข Environment` : 'เพิ่ม Environment ใหม่'}
                </h2>
                <p className="mt-1 text-sm text-slate-500">กำหนดค่า environment ให้พร้อมสำหรับการตรวจสิทธิ์</p>
              </div>
              <button type="button" onClick={closeModal} className="btn-ghost h-10 w-10 rounded-xl p-0 text-xl leading-none">
                ×
              </button>
            </div>
            <div className="modal-body">
              <form onSubmit={submit} className="space-y-4">
                {err && <div className="error-alert">{err}</div>}

                <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                  <div>
                    <label className="label">Application *</label>
                    <select
                      className="input"
                      value={form.appId}
                      onChange={(e) => setForm({ ...form, appId: Number(e.target.value) })}
                      disabled={!!editing}
                    >
                      <option value={0}>เลือก Application</option>
                      {apps.map((app) => (
                        <option key={app.id} value={app.id}>
                          {app.code} — {app.name}
                        </option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="label">Env Code *</label>
                    <input
                      className="input font-mono uppercase"
                      value={form.envCode}
                      maxLength={20}
                      onChange={(e) => setForm({ ...form, envCode: e.target.value.toUpperCase() })}
                    />
                  </div>
                  <div className="md:col-span-2">
                    <label className="label">Env Name *</label>
                    <input className="input" value={form.envName} onChange={(e) => setForm({ ...form, envName: e.target.value })} />
                  </div>
                  <div className="md:col-span-2">
                    <label className="label">Base URL *</label>
                    <input
                      className="input font-mono"
                      value={form.baseUrl}
                      placeholder="http://10.0.32.71/HelpDesk/"
                      onChange={(e) => setForm({ ...form, baseUrl: e.target.value })}
                    />
                  </div>
                  <div>
                    <label className="label">Host IP</label>
                    <input
                      className="input font-mono"
                      value={form.hostIp}
                      placeholder="10.0.32.71"
                      onChange={(e) => setForm({ ...form, hostIp: e.target.value })}
                    />
                  </div>
                  <div>
                    <label className="label">Base Path</label>
                    <input
                      className="input font-mono"
                      value={form.basePath}
                      placeholder="/HelpDesk/"
                      onChange={(e) => setForm({ ...form, basePath: e.target.value })}
                    />
                  </div>
                  <div>
                    <label className="label">AD User</label>
                    <input
                      className="input font-mono"
                      value={form.adUser}
                      maxLength={15}
                      onChange={(e) => setForm({ ...form, adUser: e.target.value })}
                    />
                  </div>
                  <label className="inline-flex items-center gap-3 rounded-xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-700">
                    <input type="checkbox" checked={form.active} onChange={(e) => setForm({ ...form, active: e.target.checked })} />
                    <span>เปิดใช้งาน Environment นี้</span>
                  </label>
                </div>

                <div className="flex justify-end gap-3 border-t border-slate-100 pt-5">
                  <button type="button" className="btn-secondary" onClick={closeModal}>ยกเลิก</button>
                  <button type="submit" className="btn-primary" disabled={saving}>
                    {saving ? 'กำลังบันทึก…' : 'บันทึก'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Modal */}
      {deleteTarget && (
        <div className="modal-overlay" onClick={() => setDeleteTarget(null)}>
          <div className="modal-content modal-animate max-w-md" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <div>
                <h2 className="text-lg font-semibold text-slate-900">ยืนยันการลบ</h2>
                <p className="mt-1 text-sm text-slate-500">
                  คุณต้องการลบ Environment <strong className="text-slate-700">{deleteTarget.envCode}</strong> ใช่หรือไม่?
                </p>
              </div>
              <button type="button" onClick={() => setDeleteTarget(null)} className="btn-ghost h-10 w-10 rounded-xl p-0 text-xl leading-none">
                ×
              </button>
            </div>
            <div className="modal-body">
              <div className="flex justify-end gap-3">
                <button type="button" className="btn-secondary" onClick={() => setDeleteTarget(null)}>ยกเลิก</button>
                <button type="button" className="btn-danger" onClick={doDelete}>ลบเลย</button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function Spinner() {
  return (
    <svg className="h-4 w-4 animate-spin text-slate-400" viewBox="0 0 24 24" fill="none">
      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
    </svg>
  )
}
