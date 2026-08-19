import { useEffect, useMemo, useRef, useState } from 'react'
import EnvManager from '../components/EnvManager.jsx'
import AppManager from '../components/AppManager.jsx'
import AllowedIpManager from '../components/AllowedIpManager.jsx'
import AllowedUserManager from '../components/AllowedUserManager.jsx'
import AuditLog from '../components/AuditLog.jsx'
import CheckAccess from '../components/CheckAccess.jsx'
import { DASHBOARD_TABS, getTabMeta } from '../navigation.js'
import { useEmployee } from '../hooks/useEmployee.js'

export default function Dashboard({ activeTab, user, onLogout, onNavigate }) {
  const [menuOpen, setMenuOpen] = useState(false)
  const [sidebarOpen, setSidebarOpen] = useState(() => {
    const saved = localStorage.getItem('sso_sidebar_open')
    return saved !== 'false' // default open
  })
  const menuRef = useRef(null)
  const tabMeta = useMemo(() => getTabMeta(activeTab), [activeTab])

  // ดึงชื่อ-นามสกุลจาก external API (proxy ผ่าน backend)
  const { employee, loading: empLoading } = useEmployee(user?.username)

  // ลำดับการแสดงชื่อ: fullName (form_first_name + form_last_name) → displayName → username
  const firstName = employee?.form_first_name || ''
  const lastName = employee?.form_last_name || ''
  const displayName =
    employee?.fullName ||
    (firstName || lastName ? `${firstName} ${lastName}`.trim() : null) ||
    user?.displayName ||
    user?.username ||
    '-'

  useEffect(() => {
    const onPointerDown = (event) => {
      if (menuRef.current && !menuRef.current.contains(event.target)) {
        setMenuOpen(false)
      }
    }
    window.addEventListener('pointerdown', onPointerDown)
    return () => window.removeEventListener('pointerdown', onPointerDown)
  }, [])

  useEffect(() => {
    localStorage.setItem('sso_sidebar_open', String(sidebarOpen))
  }, [sidebarOpen])

  const toggleSidebar = () => setSidebarOpen((prev) => !prev)

  return (
    <div className="flex h-screen overflow-hidden bg-slate-50">
      {/* Sidebar */}
      <aside
        className={
          'flex-shrink-0 border-r border-slate-200/80 bg-white flex flex-col transition-all duration-300 ease-in-out overflow-hidden ' +
          (sidebarOpen ? 'w-64' : 'w-0 border-r-0')
        }
      >
        <div className="flex items-center gap-3 px-6 h-20 border-b border-slate-200/80 min-w-[256px]">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-brand-600 text-white shadow-md shadow-brand-500/20">
            <LockIcon className="h-5 w-5" />
          </div>
          <div>
            <div className="text-base font-bold text-slate-900 leading-tight">SSO Login</div>
            <div className="text-[10px] uppercase tracking-wider text-slate-400">Controller</div>
          </div>
        </div>

        <nav className="flex-1 overflow-y-auto p-3 space-y-1 min-w-[256px]">
          {DASHBOARD_TABS.map((tab) => {
            const isActive = activeTab === tab.key
            return (
              <button
                key={tab.key}
                type="button"
                onClick={() => onNavigate(tab.key)}
                className={
                  'flex w-full items-center gap-3 px-4 py-3 text-sm font-medium transition-all duration-200 rounded-xl ' +
                  (isActive
                    ? 'bg-brand-50 text-brand-700 shadow-sm'
                    : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900')
                }
              >
                <NavIcon name={tab.icon} className="h-5 w-5" />
                <span>{tab.label}</span>
              </button>
            )
          })}
        </nav>
      </aside>

      {/* Main Content */}
      <main className="flex-1 flex flex-col min-w-0 bg-slate-50">
        {/* Navbar */}
        <header className="h-20 border-b border-slate-200/80 bg-white flex items-center justify-between px-6 flex-shrink-0">
          <div className="flex items-center gap-4">
            <button
              type="button"
              onClick={toggleSidebar}
              className="flex h-10 w-10 items-center justify-center rounded-xl border border-slate-200 bg-white text-slate-600 hover:bg-slate-50 hover:text-slate-900 transition-all duration-200"
              title={sidebarOpen ? 'ซ่อนเมนู' : 'แสดงเมนู'}
            >
              {sidebarOpen ? (
                <svg className="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <line x1="3" y1="6" x2="21" y2="6" />
                  <line x1="3" y1="12" x2="15" y2="12" />
                  <line x1="3" y1="18" x2="21" y2="18" />
                </svg>
              ) : (
                <svg className="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <line x1="3" y1="6" x2="21" y2="6" />
                  <line x1="3" y1="12" x2="21" y2="12" />
                  <line x1="3" y1="18" x2="21" y2="18" />
                </svg>
              )}
            </button>
            <div>
              <h1 className="text-xl font-bold text-slate-800">{tabMeta.label}</h1>
              <p className="text-sm text-slate-500 mt-0.5">{tabMeta.description}</p>
            </div>
          </div>

          <div className="flex items-center gap-4">
            <div ref={menuRef} className="relative">
              <button
                type="button"
                onClick={() => setMenuOpen((open) => !open)}
                className="flex items-center gap-3 bg-slate-50 px-4 py-2.5 border border-slate-200 rounded-xl hover:bg-slate-100 transition-all duration-200"
              >
                <Avatar username={user?.username} name={displayName} size="sm" />
                <div className="hidden text-left md:block">
                  <div className="text-sm font-semibold text-slate-900 leading-tight">
                    {empLoading ? (
                      <span className="inline-block h-3 w-24 bg-slate-200 rounded animate-pulse" />
                    ) : (
                      displayName
                    )}
                  </div>
                  <div className="text-[10px] text-slate-400 font-mono">
                    {user?.username || '-'}
                  </div>
                </div>
                <ChevronDownIcon className="h-4 w-4 text-slate-400" />
              </button>

              {menuOpen && (
                <div className="absolute right-0 mt-2 w-72 border border-slate-200/80 bg-white rounded-xl shadow-lg z-50 overflow-hidden">
                  <div className="bg-slate-50 p-4 border-b border-slate-100">
                    <div className="flex items-center gap-3">
                      <Avatar username={user?.username} name={displayName} size="lg" />
                      <div className="flex-1 min-w-0">
                        <div className="text-sm font-semibold text-slate-900 truncate">
                          {empLoading ? (
                            <span className="inline-block h-3 w-32 bg-slate-200 rounded animate-pulse" />
                          ) : (
                            displayName
                          )}
                        </div>
                        <div className="text-xs text-slate-500 mt-0.5 truncate">
                          ADUser: {user?.username || '-'}
                        </div>
                      </div>
                    </div>
                    <div className="mt-3 flex items-center gap-2">
                      <span
                        className={
                          'inline-flex items-center px-2 py-0.5 rounded-md text-[10px] font-semibold uppercase tracking-wider ' +
                          (user?.role === 'admin'
                            ? 'bg-purple-100 text-purple-700'
                            : 'bg-blue-100 text-blue-700')
                        }
                      >
                        {user?.role === 'admin' ? 'Admin' : 'User'}
                      </span>
                    </div>
                  </div>

                  <div className="p-2">
                    <button
                      type="button"
                      onClick={onLogout}
                      className="flex w-full items-center gap-2 px-4 py-2.5 text-sm font-medium text-rose-600 hover:bg-rose-50 rounded-lg transition-colors"
                    >
                      <LogoutIcon className="h-4 w-4" />
                      ออกจากระบบ
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>
        </header>

        {/* Content Area */}
        <div className="flex-1 overflow-auto p-6 md:p-8">
          <div className="w-full">
            {activeTab === 'envs' && <EnvManager />}
            {activeTab === 'apps' && <AppManager />}
            {activeTab === 'ips' && <AllowedIpManager />}
            {activeTab === 'users' && <AllowedUserManager />}
            {activeTab === 'check' && <CheckAccess />}
            {activeTab === 'audit' && <AuditLog />}
          </div>
        </div>

        {/* Footer */}
        <footer className="h-12 border-t border-slate-200 bg-white flex items-center justify-between px-8 text-xs text-slate-500 flex-shrink-0">
          <div>
            &copy; {new Date().getFullYear()} SSO Login System
          </div>
          <div className="flex items-center gap-2">
            <span className="font-semibold text-slate-700">User:</span> {displayName}
            <span className="mx-2 text-slate-300">|</span>
            <span className="font-semibold text-slate-700">Session expires:</span> {formatDateTime(user?.expiresAt)} ({formatRelativeExpiry(user?.expiresAt)})
          </div>
        </footer>
      </main>
    </div>
  )
}

function formatDateTime(value) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return date.toLocaleString('th-TH', {
    dateStyle: 'medium',
    timeStyle: 'short',
  })
}

function formatRelativeExpiry(value) {
  if (!value) return '-'
  const expiry = new Date(value)
  if (Number.isNaN(expiry.getTime())) return '-'
  const diffMinutes = Math.max(0, Math.round((expiry.getTime() - Date.now()) / 60000))
  if (diffMinutes >= 24 * 60) {
    const days = Math.floor(diffMinutes / (24 * 60))
    return `อีก ${days} วัน`
  }
  if (diffMinutes >= 60) {
    const hours = Math.floor(diffMinutes / 60)
    return `อีก ${hours} ชม.`
  }
  return `อีก ${diffMinutes} นาที`
}

function NavIcon({ name, className }) {
  if (name === 'layers') {
    return (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" className={className}>
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M12 3l9 4.5-9 4.5L3 7.5 12 3z" />
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M3 12l9 4.5L21 12" />
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M3 16.5L12 21l9-4.5" />
      </svg>
    )
  }
  if (name === 'grid') {
    return (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" className={className}>
        <rect x="3" y="3" width="7" height="7" rx="1.5" strokeWidth="1.8" />
        <rect x="14" y="3" width="7" height="7" rx="1.5" strokeWidth="1.8" />
        <rect x="3" y="14" width="7" height="7" rx="1.5" strokeWidth="1.8" />
        <rect x="14" y="14" width="7" height="7" rx="1.5" strokeWidth="1.8" />
      </svg>
    )
  }
  if (name === 'globe') {
    return (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" className={className}>
        <circle cx="12" cy="12" r="9" strokeWidth="1.8" />
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M3 12h18M12 3c2.5 2.7 4 5.7 4 9s-1.5 6.3-4 9c-2.5-2.7-4-5.7-4-9s1.5-6.3 4-9z" />
      </svg>
    )
  }
  if (name === 'users') {
    return (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" className={className}>
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M16 21v-1.4A3.6 3.6 0 0012.4 16h-4.8A3.6 3.6 0 004 19.6V21" />
        <circle cx="10" cy="8" r="3" strokeWidth="1.8" />
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M20 21v-1.6A3.4 3.4 0 0017 16.1M14.5 4.8A3 3 0 0117 10" />
      </svg>
    )
  }
  if (name === 'shield') {
    return (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" className={className}>
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M12 3l7 3v5c0 4.5-2.7 7.8-7 10-4.3-2.2-7-5.5-7-10V6l7-3z" />
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M9.5 12.5l1.7 1.7 3.6-4" />
      </svg>
    )
  }
  if (name === 'receipt') {
    return <ReceiptIcon className={className} />
  }
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" className={className}>
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M4 6.5h16M4 12h16M4 17.5h16" />
    </svg>
  )
}

function LockIcon({ className }) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" className={className}>
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M7 10V8a5 5 0 0110 0v2" />
      <rect x="5" y="10" width="14" height="11" rx="2.5" strokeWidth="1.8" />
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M12 14.5v2.5" />
    </svg>
  )
}

function UserIcon({ className }) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" className={className}>
      <circle cx="12" cy="8" r="3.5" strokeWidth="1.8" />
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M5 19.5c1.8-2.5 4.2-3.8 7-3.8s5.2 1.3 7 3.8" />
    </svg>
  )
}

// สร้าง URL รูป avatar จาก ADUser
// เช่น T9058 -> base64 = "VDkwNTg=" -> ตัด 1 ตัวท้าย -> "VDkwNTg"
// URL: http://api.tnlx.co.th/hr/images_HR/VDkwNTg.jpg?152607
const AVATAR_BASE = 'http://api.tnlx.co.th/hr/images_HR'
const AVATAR_QUERY = '?152607'
function getAvatarUrl(username) {
  if (!username) return null
  try {
    // btoa รองรับ ASCII เท่านั้น — ADUser (เช่น T9058) เป็น ASCII อยู่แล้ว
    const b64 = btoa(username)
    const trimmed = b64.slice(0, -1) // ตัดตัวสุดท้ายออก 1 ตัว
    return `${AVATAR_BASE}/${trimmed}.jpg${AVATAR_QUERY}`
  } catch {
    return null
  }
}

// Avatar — แสดงรูปจาก ADUser, fallback ไป icon user ถ้าโหลดไม่สำเร็จ
function Avatar({ username, name, size = 'sm' }) {
  const [errored, setErrored] = useState(false)
  const url = getAvatarUrl(username)
  const sizeCls = size === 'lg' ? 'h-10 w-10' : 'h-9 w-9'
  const iconCls = size === 'lg' ? 'h-5 w-5' : 'h-4 w-4'
  if (!url || errored) {
    return (
      <div className={`flex ${sizeCls} items-center justify-center rounded-full bg-slate-200 text-slate-600`}>
        <UserIcon className={iconCls} />
      </div>
    )
  }
  return (
    <div className={`flex ${sizeCls} items-center justify-center rounded-full bg-slate-200 overflow-hidden flex-shrink-0`}>
      <img
        src={url}
        alt={name || username}
        className={`${sizeCls} object-cover`}
        onError={() => setErrored(true)}
      />
    </div>
  )
}

function ChevronDownIcon({ className }) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" className={className}>
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M6 9l6 6 6-6" />
    </svg>
  )
}

function LogoutIcon({ className }) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" className={className}>
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M15 16l4-4-4-4" />
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M19 12H9" />
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M11 19H6a2 2 0 01-2-2V7a2 2 0 012-2h5" />
    </svg>
  )
}

function ReceiptIcon({ className }) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" className={className}>
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M7 3h10v18l-2.5-1.8L12 21l-2.5-1.8L7 21V3z" />
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M9.5 8h5M9.5 12h5M9.5 16h3.5" />
    </svg>
  )
}
