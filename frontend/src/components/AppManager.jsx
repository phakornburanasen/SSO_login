import { useEffect, useMemo, useState } from 'react'
import { api } from '../api.js'
import Pagination from './Pagination.jsx'

const EMPTY = { code: '', name: '', description: '', active: true }

export default function AppManager() {
  const [items, setItems] = useState([])
  const [form, setForm] = useState(EMPTY)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState(null)
  const [err, setErr] = useState('')
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
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
      const r = await api.listApps()
      setItems(r.apps || [])
    } catch (e) {
      setErr(e.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  useEffect(() => {
    const totalPages = Math.max(1, Math.ceil(items.length / pageSize))
    if (page > totalPages) setPage(totalPages)
  }, [items.length, page, pageSize])

  const openNew = () => {
    setEditing(null)
    setForm(EMPTY)
    setErr('')
    setModalOpen(true)
  }

  const openEdit = (row) => {
    setEditing(row)
    setForm({
      code: row.code,
      name: row.name,
      description: row.description,
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
    if (!form.code.trim() || !form.name.trim()) return setErr('กรุณากรอก Code และ Name')
    setSaving(true)
    try {
      if (editing) await api.updateApp(editing.id, form)
      else await api.createApp(form)
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
      await api.deleteApp(deleteTarget.id)
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
          <h2 className="section-title">รายการแอปพลิเคชัน</h2>
          <p className="section-desc">กำหนด code, ชื่อระบบ และสถานะการใช้งานของแต่ละ application</p>
        </div>
        <div className="ml-auto">
          <button type="button" className="btn-primary" onClick={openNew}>
            + เพิ่ม Application
          </button>
        </div>
      </div>

      {err && <div className="error-alert fade-in">{err}</div>}

      <div className="table-shell">
        <div className="overflow-x-auto">
          <table className="table">
            <thead>
              <tr>
                <th className="th">Code</th>
                <th className="th">Name</th>
                <th className="th">Description</th>
                <th className="th">Status</th>
                <th className="th w-[180px]">จัดการ</th>
              </tr>
            </thead>
            <tbody>
              {loading && (
                <tr>
                  <td colSpan={5} className="td py-10 text-center text-slate-400">
                    <div className="flex items-center justify-center gap-2">
                      <Spinner /> กำลังโหลดข้อมูล…
                    </div>
                  </td>
                </tr>
              )}
              {!loading && items.length === 0 && (
                <tr>
                  <td colSpan={5} className="td py-10 text-center text-slate-400">ยังไม่มี Application ในระบบ</td>
                </tr>
              )}
              {!loading &&
                pagedItems.map((row) => (
                  <tr key={row.id} className="table-row-hover fade-in">
                    <td className="td">
                      <span className="inline-flex rounded-lg bg-slate-100 px-2.5 py-1 font-mono text-xs font-semibold text-slate-700">
                        {row.code}
                      </span>
                    </td>
                    <td className="td font-medium text-slate-900">{row.name}</td>
                    <td className="td text-slate-500">{row.description || '-'}</td>
                    <td className="td">
                      {row.active ? (
                        <span className="badge-on">Active</span>
                      ) : (
                        <span className="badge-off">Inactive</span>
                      )}
                    </td>
                    <td className="td">
                      <div className="flex gap-2">
                        <button
                          type="button"
                          className="btn-secondary px-3 py-1.5 text-xs"
                          onClick={() => openEdit(row)}
                        >
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
          totalItems={items.length}
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
                  {editing ? `แก้ไข Application` : 'เพิ่ม Application ใหม่'}
                </h2>
                <p className="mt-1 text-sm text-slate-500">
                  {editing
                    ? `แก้ไขข้อมูลของ ${editing.code}`
                    : 'กำหนดข้อมูลหลักของระบบให้พร้อมใช้งาน'}
                </p>
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
                    <label className="label">Code *</label>
                    <input
                      className="input font-mono uppercase"
                      value={form.code}
                      disabled={!!editing}
                      onChange={(e) => setForm({ ...form, code: e.target.value.toUpperCase() })}
                    />
                  </div>
                  <div>
                    <label className="label">Name *</label>
                    <input
                      className="input"
                      value={form.name}
                      onChange={(e) => setForm({ ...form, name: e.target.value })}
                    />
                  </div>
                  <div className="md:col-span-2">
                    <label className="label">Description</label>
                    <textarea
                      className="input min-h-[100px] resize-y"
                      value={form.description}
                      onChange={(e) => setForm({ ...form, description: e.target.value })}
                    />
                  </div>
                  <label className="inline-flex items-center gap-3 rounded-xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-700">
                    <input
                      type="checkbox"
                      checked={form.active}
                      onChange={(e) => setForm({ ...form, active: e.target.checked })}
                    />
                    <span>เปิดใช้งาน Application นี้</span>
                  </label>
                </div>

                <div className="flex justify-end gap-3 border-t border-slate-100 pt-5">
                  <button type="button" className="btn-secondary" onClick={closeModal}>
                    ยกเลิก
                  </button>
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
                  คุณต้องการลบ Application <strong className="text-slate-700">{deleteTarget.code}</strong> ใช่หรือไม่?
                </p>
              </div>
              <button type="button" onClick={() => setDeleteTarget(null)} className="btn-ghost h-10 w-10 rounded-xl p-0 text-xl leading-none">
                ×
              </button>
            </div>
            <div className="modal-body">
              <div className="flex justify-end gap-3">
                <button type="button" className="btn-secondary" onClick={() => setDeleteTarget(null)}>
                  ยกเลิก
                </button>
                <button type="button" className="btn-danger" onClick={doDelete}>
                  ลบเลย
                </button>
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
