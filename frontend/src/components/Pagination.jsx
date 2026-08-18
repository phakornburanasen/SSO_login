export default function Pagination({
  page,
  totalItems,
  pageSize,
  onPageChange,
  pageSizeOptions = [5, 10, 20, 50],
  onPageSizeChange,
}) {
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize))
  const start = totalItems === 0 ? 0 : (page - 1) * pageSize + 1
  const end = Math.min(totalItems, page * pageSize)

  return (
    <div className="flex flex-col gap-3 border-t border-slate-100 bg-slate-50/50 px-5 py-4 md:flex-row md:items-center md:justify-between">
      <div className="text-sm text-slate-500">
        แสดง <span className="font-medium text-slate-700">{start}-{end}</span> จาก{' '}
        <span className="font-medium text-slate-700">{totalItems}</span> รายการ
      </div>

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
        {onPageSizeChange && (
          <label className="flex items-center gap-2 text-sm text-slate-500">
            <span>ต่อหน้า</span>
            <select
              className="input min-w-[88px] py-2"
              value={pageSize}
              onChange={(event) => onPageSizeChange(Number(event.target.value))}
            >
              {pageSizeOptions.map((size) => (
                <option key={size} value={size}>
                  {size}
                </option>
              ))}
            </select>
          </label>
        )}

        <div className="flex items-center gap-2">
          <button
            type="button"
            className="btn-secondary px-3 py-2"
            disabled={page <= 1}
            onClick={() => onPageChange(page - 1)}
          >
            ก่อนหน้า
          </button>
          <div className="rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm font-medium text-slate-700 shadow-sm">
            {page} / {totalPages}
          </div>
          <button
            type="button"
            className="btn-secondary px-3 py-2"
            disabled={page >= totalPages}
            onClick={() => onPageChange(page + 1)}
          >
            ถัดไป
          </button>
        </div>
      </div>
    </div>
  )
}
