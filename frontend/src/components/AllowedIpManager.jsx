import { useEffect, useMemo, useState } from 'react'
import { api } from '../api.js'
import Pagination from './Pagination.jsx'

const EMPTY = { envId: 0, ipCidr: '', description: '' }

export default function AllowedIpManager() {
  const [envs, setEnvs] = useState([])
  const [items, setItems] = useState([])
  const [filterEnv, setFilterEnv] = useState(0)
  const [form, setForm] = useState(EMPTY)
  const [err, setErr] = useState('')
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [deleteTarget, setDeleteTarget] = useState(null)

  const pagedItems = useMemo(() => {
    const start = (page - 1) * pageSize
    return items.slice(start, start + pageSize)
  }, [items, page, pageSize])

  const load = async () => {
    setLoading(true)
    setErr('')
    try {
      const [e, i] = await Promise.all([api.listEnvs(0), api.listAllowedIPs(filterEnv)])
      setEnvs(e.envs || [])
      setItems(i.allowedIps || [])
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

  const openModal = () => {
    setForm({ ...EMPTY, envId: filterEnv || envs[0]?.id || 0 })
    setErr('')
    setModalOpen(true)
  }

  const closeModal = () => {
    setModalOpen(false)
    setForm(EMPTY)
    setErr('')
  }

  const submit = async (e) => {
    e.preventDefault()
    setErr('')
    if (!form.envId || !form.ipCidr.trim()) return setErr('กรุณาเลือก Environment และกรอก IP/CIDR')
    setSaving(true)
    try {
      await api.createAllowedIP({ ...form, createdBy: 'admin-ui' })
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
      await api.deleteAllowedIP(deleteTarget.id)
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
          <h2 className="section-title">Allowed IP / CIDR</h2>
          <p className="section-desc">จัดการเครือข่ายที่อนุญาตให้เข้าแต่ละ environment ผ่าน policy ของระบบ</p>
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
          <button type="button" className="btn-primary" onClick={openModal}>
            + เพิ่ม Allowed IP
          </button>
        </div>
      </div>

      {err && <div className="error-alert fade-in">{err}</div>}

      <div className="table-shell">
        <div className="overflow-x-auto">
          <table className="table">
            <thead>
              <tr>
                <th className="th">Environment</th>
                <th className="th">IP / CIDR</th>
                <th className="th">Description</th>
                <th className="th">Status</th>
                <th className="th">Created By</th>
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
                  <td colSpan={6} className="td py-10 text-center text-slate-400">ยังไม่มี Allowed IP</td>
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
                      <td className="td font-mono text-sm text-slate-900">{row.ipCidr}</td>
                      <td className="td text-slate-500">{row.description || '-'}</td>
                      <td className="td">
                        {row.active ? <span className="badge-on">Active</span> : <span className="badge-off">Off</span>}
                      </td>
                      <td className="td text-sm text-slate-500">{row.createdBy || '-'}</td>
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

      {/* Add Modal */}
      {modalOpen && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal-content modal-animate max-w-xl" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <div>
                <h2 className="text-lg font-semibold text-slate-900">เพิ่ม Allowed IP</h2>
                <p className="mt-1 text-sm text-slate-500">กำหนดเครือข่ายที่อนุญาตให้เข้าถึง environment</p>
              </div>
              <button type="button" onClick={closeModal} className="btn-ghost h-10 w-10 rounded-xl p-0 text-xl leading-none">×</button>
            </div>
            <div className="modal-body">
              <form onSubmit={submit} className="space-y-4">
                {err && <div className="error-alert">{err}</div>}
                <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                  <div>
                    <label className="label">Environment *</label>
                    <select className="input" value={form.envId} onChange={(e) => setForm({ ...form, envId: Number(e.target.value) })}>
                      <option value={0}>เลือก Environment</option>
                      {envs.map((en) => (
                        <option key={en.id} value={en.id}>{en.appCode}/{en.envCode}</option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="label">IP / CIDR *</label>
                    <input className="input font-mono" value={form.ipCidr} placeholder="10.0.32.0/24" onChange={(e) => setForm({ ...form, ipCidr: e.target.value })} />
                  </div>
                  <div className="md:col-span-2">
                    <label className="label">Description</label>
                    <input className="input" value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
                  </div>
                </div>
                <div className="flex justify-end gap-3 border-t border-slate-100 pt-5">
                  <button type="button" className="btn-secondary" onClick={closeModal}>ยกเลิก</button>
                  <button type="submit" className="btn-primary" disabled={saving}>{saving ? 'กำลังบันทึก…' : 'บันทึก'}</button>
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
                <p className="mt-1 text-sm text-slate-500">คุณต้องการลบ IP <strong className="text-slate-700">{deleteTarget.ipCidr}</strong> ใช่หรือไม่?</p>
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

function Spinner() {
  return (
    <svg className="h-4 w-4 animate-spin text-slate-400" viewBox="0 0 24 24" fill="none">
      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
    </svg>
  )
}
