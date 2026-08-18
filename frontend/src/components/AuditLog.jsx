import { useEffect, useMemo, useState } from 'react'
import { api } from '../api.js'
import Pagination from './Pagination.jsx'

export default function AuditLog() {
  const [items, setItems] = useState([])
  const [err, setErr] = useState('')
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)

  const pagedItems = useMemo(() => {
    const start = (page - 1) * pageSize
    return items.slice(start, start + pageSize)
  }, [items, page, pageSize])

  const load = async () => {
    setLoading(true)
    setErr('')
    try {
      const r = await api.listAudit(200)
      setItems(r.audit || [])
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

  return (
    <div className="space-y-6">
      <div className="table-toolbar">
        <div>
          <h2 className="section-title">Audit Log</h2>
          <p className="section-desc">ติดตามผลการตรวจสิทธิ์ย้อนหลัง พร้อมเหตุผลการ allow หรือ deny ในแต่ละ request</p>
        </div>
        <div className="ml-auto">
          <button type="button" className="btn-secondary" onClick={load} disabled={loading}>
            รีเฟรชข้อมูล
          </button>
        </div>
      </div>

      {err && <div className="error-alert fade-in">{err}</div>}

      <div className="table-shell">
        <div className="overflow-x-auto">
          <table className="table">
            <thead>
              <tr>
                <th className="th">เวลา</th>
                <th className="th">App/Env</th>
                <th className="th">Base URL</th>
                <th className="th">Client IP</th>
                <th className="th">AD User</th>
                <th className="th">Result</th>
                <th className="th">Reason</th>
              </tr>
            </thead>
            <tbody>
              {loading && (
                <tr>
                  <td colSpan={7} className="td py-10 text-center text-slate-400">
                    <div className="flex items-center justify-center gap-2">
                      <Spinner /> กำลังโหลดข้อมูล…
                    </div>
                  </td>
                </tr>
              )}
              {!loading && items.length === 0 && (
                <tr>
                  <td colSpan={7} className="td py-10 text-center text-slate-400">ยังไม่มี Audit Log</td>
                </tr>
              )}
              {!loading &&
                pagedItems.map((row) => (
                  <tr key={row.id} className="table-row-hover fade-in">
                    <td className="td text-sm text-slate-500">{formatDateTime(row.createdAt)}</td>
                    <td className="td text-xs">
                      <span className="inline-flex rounded-lg bg-slate-100 px-2.5 py-1 font-mono font-semibold text-slate-700">
                        {row.appCode}/{row.envCode}
                      </span>
                    </td>
                    <td className="td max-w-[340px] break-all font-mono text-xs text-slate-600">{row.baseUrl || '-'}</td>
                    <td className="td font-mono text-sm">{row.clientIp || '-'}</td>
                    <td className="td font-mono text-sm">{row.adUsername || '-'}</td>
                    <td className="td">
                      <span className={resultClass(row.result)}>{row.result}</span>
                    </td>
                    <td className="td text-sm text-slate-500">{row.denyReason || '-'}</td>
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
    </div>
  )
}

function resultClass(result) {
  if (result === 'ALLOW') return 'badge-on'
  if (result && result.startsWith('DENY')) return 'badge bg-rose-50 text-rose-600 ring-1 ring-rose-200/60'
  return 'badge bg-amber-50 text-amber-700 ring-1 ring-amber-200/60'
}

function formatDateTime(value) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return date.toLocaleString('th-TH', { dateStyle: 'medium', timeStyle: 'short' })
}

function Spinner() {
  return (
    <svg className="h-4 w-4 animate-spin text-slate-400" viewBox="0 0 24 24" fill="none">
      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
    </svg>
  )
}
