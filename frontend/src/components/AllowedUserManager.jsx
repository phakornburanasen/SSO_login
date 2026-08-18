import { useEffect, useMemo, useState } from 'react'
import { api } from '../api.js'
import Pagination from './Pagination.jsx'

export default function AllowedUserManager() {
  const [envs, setEnvs] = useState([])
  const [apps, setApps] = useState([])
  const [items, setItems] = useState([])
  const [filterEnv, setFilterEnv] = useState(0)
  const [err, setErr] = useState('')
  const [success, setSuccess] = useState('')
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [deleteTarget, setDeleteTarget] = useState(null)

  // Grant form state
  const [adUsername, setAdUsername] = useState('')
  const [selectedEnvIds, setSelectedEnvIds] = useState([])
  const [expiresAt, setExpiresAt] = useState('')
  const [saving, setSaving] = useState(false)

  // Copy permissions state
  const [copyFromUser, setCopyFromUser] = useState('')
  const [copyFromEnvs, setCopyFromEnvs] = useState([])
  const [loadingCopy, setLoadingCopy] = useState(false)

  const pagedItems = useMemo(() => {
    const start = (page - 1) * pageSize
    return items.slice(start, start + pageSize)
  }, [items, page, pageSize])

  const load = async () => {
    setLoading(true)
    setErr('')
    try {
      const [a, e, u] = await Promise.all([
        api.listApps(),
        api.listEnvs(0),
        api.listAllowedUsers(filterEnv),
      ])
      setApps(a.apps || [])
      setEnvs(e.envs || [])
      setItems(u.allowedUsers || [])
    } catch (e) {
      setErr(e.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [filterEnv])

  useEffect(() => {
    const totalPages = Math.max(1, Math.ceil(items.length / pageSize))
    if (page > totalPages) setPage(totalPages)
  }, [items.length, page, pageSize])

  // Group envs by app for checkbox display
  const envsByApp = useMemo(() => {
    const groups = {}
    for (const env of envs) {
      if (!groups[env.appCode]) groups[env.appCode] = { appName: '', envs: [] }
      groups[env.appCode].appName = env.appCode
      groups[env.appCode].envs.push(env)
    }
    return Object.values(groups)
  }, [envs])

  const toggleEnv = (envId) => {
    setSelectedEnvIds((prev) =>
      prev.includes(envId) ? prev.filter((id) => id !== envId) : [...prev, envId]
    )
  }

  const selectAllEnvs = () => {
    setSelectedEnvIds(envs.map((e) => e.id))
  }

  const deselectAllEnvs = () => {
    setSelectedEnvIds([])
  }

  const handleCopyPermissions = async () => {
    if (!copyFromUser.trim()) {
      setErr('กรุณาใส่ ADUser ที่ต้องการอ้างอิงสิทธิ์')
      return
    }
    setLoadingCopy(true)
    setErr('')
    try {
      const res = await api.listAllowedUsersByUser(copyFromUser.trim())
      const users = res.allowedUsers || []
      if (users.length === 0) {
        setErr(`ไม่พบสิทธิ์ของ "${copyFromUser.trim()}" ในระบบ`)
        setCopyFromEnvs([])
        return
      }
      setCopyFromEnvs(users)
      // Auto-select envs from the reference user
      setSelectedEnvIds(users.map((u) => u.envId))
      setSuccess(`ดึงสิทธิ์จาก "${copyFromUser.trim()}" สำเร็จ — ${users.length} ระบบ`)
    } catch (e) {
      setErr(e.message)
    } finally {
      setLoadingCopy(false)
    }
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setErr('')
    setSuccess('')
    if (!adUsername.trim()) return setErr('กรุณากรอก ADUser')
    if (selectedEnvIds.length === 0) return setErr('กรุณาเลือกระบบอย่างน้อย 1 ระบบ')

    setSaving(true)
    try {
      const body = {
        adUsername: adUsername.trim(),
        envIds: selectedEnvIds,
      }
      if (expiresAt) body.expiresAt = expiresAt
      const res = await api.bulkCreateAllowedUsers(body)
      setSuccess(
        `เพิ่มสิทธิ์ให้ "${adUsername.trim()}" สำเร็จ — ${res.created?.length || 0} ระบบ` +
        (res.skipped?.length > 0 ? ` (ข้าม ${res.skipped.length} ระบบที่มีอยู่แล้ว)` : '')
      )
      setAdUsername('')
      setSelectedEnvIds([])
      setExpiresAt('')
      setCopyFromUser('')
      setCopyFromEnvs([])
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
      await api.deleteAllowedUser(deleteTarget.id)
      setDeleteTarget(null)
      await load()
    } catch (e) {
      setErr(e.message)
    }
  }

  const getEnvLabel = (envId) => {
    const env = envs.find((e) => e.id === envId)
    return env ? `${env.appCode}/${env.envCode}` : `#${envId}`
  }

  return (
    <div className="space-y-6">
      {/* Grant Form Card */}
      <div className="card p-6">
        <div className="mb-6">
          <h2 className="section-title">เพิ่มสิทธิ์ผู้ใช้งาน</h2>
          <p className="section-desc">กำหนดสิทธิ์ให้ AD User เข้าถึงหลายระบบพร้อมกันในครั้งเดียว</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-6">
          {err && <div className="error-alert fade-in">{err}</div>}
          {success && (
            <div className="rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700 fade-in">
              {success}
            </div>
          )}

          {/* ADUser */}
          <div>
            <label className="label">ADUser *</label>
            <input
              className="input font-mono"
              value={adUsername}
              onChange={(e) => setAdUsername(e.target.value)}
              placeholder="somchai.s"
            />
          </div>

          {/* Copy Permissions */}
          <div className="rounded-xl border border-slate-200 bg-slate-50/50 p-4">
            <div className="mb-3">
              <label className="label">อ้างอิงสิทธิ์จาก AD User อื่น (ไม่บังคับ)</label>
              <p className="text-xs text-slate-500">ใส่ ADUser เพื่อดูว่าผู้นั้นมีสิทธิ์เข้าระบบอะไรบ้าง แล้วติ๊กตามได้ทันที</p>
            </div>
            <div className="flex gap-3">
              <input
                className="input font-mono flex-1"
                value={copyFromUser}
                onChange={(e) => setCopyFromUser(e.target.value)}
                placeholder="ADUser ที่ต้องการอ้างอิง"
              />
              <button
                type="button"
                className="btn-secondary whitespace-nowrap"
                onClick={handleCopyPermissions}
                disabled={loadingCopy}
              >
                {loadingCopy ? (
                  <span className="flex items-center gap-2"><Spinner /> กำลังดึงข้อมูล…</span>
                ) : (
                  'ดึงสิทธิ์'
                )}
              </button>
            </div>
            {copyFromEnvs.length > 0 && (
              <div className="mt-3 flex flex-wrap gap-1.5">
                {copyFromEnvs.map((u) => (
                  <span key={u.envId} className="inline-flex rounded-lg bg-blue-50 px-2 py-1 text-xs font-medium text-blue-700 ring-1 ring-blue-200">
                    {getEnvLabel(u.envId)}
                  </span>
                ))}
              </div>
            )}
          </div>

          {/* System Checkboxes */}
          <div>
            <div className="mb-3 flex items-center justify-between">
              <label className="label mb-0">เลือกระบบที่ต้องการให้สิทธิ์ *</label>
              <div className="flex gap-2">
                <button type="button" className="text-xs font-medium text-brand-600 hover:text-brand-700" onClick={selectAllEnvs}>
                  เลือกทั้งหมด
                </button>
                <span className="text-slate-300">|</span>
                <button type="button" className="text-xs font-medium text-slate-500 hover:text-slate-700" onClick={deselectAllEnvs}>
                  ล้างทั้งหมด
                </button>
              </div>
            </div>

            {envs.length === 0 ? (
              <div className="rounded-xl border-2 border-dashed border-slate-200 bg-slate-50/50 px-5 py-8 text-center text-sm text-slate-400">
                ยังไม่มีระบบในฐานข้อมูล — กรุณาเพิ่ม Application และ Environment ก่อน
              </div>
            ) : (
              <div className="space-y-4">
                {envsByApp.map((group) => (
                  <div key={group.appName} className="rounded-xl border border-slate-200 bg-white p-4">
                    <div className="mb-3 text-sm font-semibold text-slate-700">{group.appName}</div>
                    <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
                      {group.envs.map((env) => {
                        const isChecked = selectedEnvIds.includes(env.id)
                        return (
                          <label
                            key={env.id}
                            className={
                              'flex cursor-pointer items-start gap-3 rounded-xl border p-3 transition-all duration-150 ' +
                              (isChecked
                                ? 'border-brand-300 bg-brand-50/50 shadow-sm'
                                : 'border-slate-200 bg-white hover:border-slate-300 hover:bg-slate-50')
                            }
                          >
                            <input
                              type="checkbox"
                              checked={isChecked}
                              onChange={() => toggleEnv(env.id)}
                              className="mt-0.5 h-4 w-4 rounded border-slate-300 text-brand-600 focus:ring-brand-500"
                            />
                            <div className="min-w-0 flex-1">
                              <div className="text-sm font-medium text-slate-900">{env.envCode}</div>
                              <div className="truncate text-xs text-slate-500">{env.envName}</div>
                              <div className="truncate font-mono text-[10px] text-slate-400">{env.baseUrl}</div>
                            </div>
                          </label>
                        )
                      })}
                    </div>
                  </div>
                ))}
              </div>
            )}

            {selectedEnvIds.length > 0 && (
              <div className="mt-3 text-sm text-slate-600">
                เลือกแล้ว <span className="font-semibold text-brand-600">{selectedEnvIds.length}</span> ระบบ
              </div>
            )}
          </div>

          {/* Expiry */}
          <div>
            <label className="label">วันหมดอายุ (ไม่บังคับ — เว้นว่าง = ไม่หมดอายุ)</label>
            <input
              type="datetime-local"
              className="input"
              value={expiresAt}
              onChange={(e) => setExpiresAt(e.target.value)}
            />
          </div>

          {/* Submit */}
          <div className="flex justify-end border-t border-slate-100 pt-5">
            <button type="submit" className="btn-primary" disabled={saving || selectedEnvIds.length === 0}>
              {saving ? (
                <span className="flex items-center gap-2"><Spinner /> กำลังบันทึก…</span>
              ) : (
                `บันทึกสิทธิ์ (${selectedEnvIds.length} ระบบ)`
              )}
            </button>
          </div>
        </form>
      </div>

      {/* Existing Grants Table */}
      <div className="table-toolbar">
        <div>
          <h2 className="section-title">สิทธิ์ที่มีอยู่ในระบบ</h2>
          <p className="section-desc">รายชื่อ AD User ที่ได้รับสิทธิ์ในแต่ละ environment</p>
        </div>
        <div className="subcard ml-auto flex flex-wrap items-end gap-3 px-4 py-4">
          <div>
            <label className="label">กรองตาม Environment</label>
            <select
              className="input min-w-[280px]"
              value={filterEnv}
              onChange={(e) => {
                setFilterEnv(Number(e.target.value))
                setPage(1)
              }}
            >
              <option value={0}>ทั้งหมด</option>
              {envs.map((en) => (
                <option key={en.id} value={en.id}>
                  {en.appCode}/{en.envCode} — {en.envName}
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

      <div className="table-shell">
        <div className="overflow-x-auto">
          <table className="table">
            <thead>
              <tr>
                <th className="th">Environment</th>
                <th className="th">ADUser</th>
                <th className="th">Display Name</th>
                <th className="th">Email</th>
                <th className="th">หมดอายุ</th>
                <th className="th w-[120px]">จัดการ</th>
              </tr>
            </thead>
            <tbody>
              {loading && (
                <tr>
                  <td colSpan={6} className="td py-10 text-center text-slate-400">
                    <div className="flex items-center justify-center gap-2">
                      <Spinner /> กำลังโหลดข้อมูล…
                    </div>
                  </td>
                </tr>
              )}
              {!loading && items.length === 0 && (
                <tr>
                  <td colSpan={6} className="td py-10 text-center text-slate-400">ยังไม่มี Allowed User</td>
                </tr>
              )}
              {!loading &&
                pagedItems.map((row) => {
                  const en = envs.find((item) => item.id === row.envId)
                  return (
                    <tr key={row.id} className="table-row-hover fade-in">
                      <td className="td text-xs">
                        <span className="inline-flex rounded-lg bg-slate-100 px-2.5 py-1 font-mono font-semibold text-slate-700">
                          {en ? `${en.appCode}/${en.envCode}` : `#${row.envId}`}
                        </span>
                      </td>
                      <td className="td font-mono text-sm font-semibold text-slate-900">{row.adUsername}</td>
                      <td className="td font-medium text-slate-900">{row.displayName || '-'}</td>
                      <td className="td text-sm text-slate-500">{row.email || '-'}</td>
                      <td className="td text-sm text-slate-500">{formatDateTime(row.expiresAt)}</td>
                      <td className="td">
                        <button
                          type="button"
                          className="btn-ghost px-3 py-1.5 text-xs text-rose-600 hover:text-rose-700"
                          onClick={() => confirmDelete(row)}
                        >
                          ลบ
                        </button>
                      </td>
                    </tr>
                  )
                })}
            </tbody>
          </table>
        </div>

        <Pagination
          page={page}
          totalItems={items.length}
          pageSize={pageSize}
          onPageChange={setPage}
          onPageSizeChange={(size) => {
            setPageSize(size)
            setPage(1)
          }}
        />
      </div>

      {/* Delete Confirmation Modal */}
      {deleteTarget && (
        <div className="modal-overlay" onClick={() => setDeleteTarget(null)}>
          <div className="modal-content modal-animate max-w-md" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <div>
                <h2 className="text-lg font-semibold text-slate-900">ยืนยันการลบ</h2>
                <p className="mt-1 text-sm text-slate-500">
                  คุณต้องการลบสิทธิ์ของ <strong className="text-slate-700">{deleteTarget.adUsername}</strong> ในระบบ{' '}
                  <strong className="text-slate-700">{getEnvLabel(deleteTarget.envId)}</strong> ใช่หรือไม่?
                </p>
              </div>
              <button type="button" onClick={() => setDeleteTarget(null)} className="btn-ghost h-10 w-10 rounded-xl p-0 text-xl leading-none">×</button>
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

function formatDateTime(value) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  
  // Format: DD/MM/YYYY hh:mm
  const day = String(date.getDate()).padStart(2, '0')
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const year = date.getFullYear()
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  
  return `${day}/${month}/${year} ${hours}:${minutes}`
}

function Spinner() {
  return (
    <svg className="h-4 w-4 animate-spin text-slate-400" viewBox="0 0 24 24" fill="none">
      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
    </svg>
  )
}
