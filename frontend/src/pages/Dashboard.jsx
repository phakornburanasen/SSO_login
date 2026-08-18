import { useState } from 'react'
import EnvManager      from '../components/EnvManager.jsx'
import AppManager      from '../components/AppManager.jsx'
import AllowedIpManager   from '../components/AllowedIpManager.jsx'
import AllowedUserManager from '../components/AllowedUserManager.jsx'
import AuditLog        from '../components/AuditLog.jsx'
import CheckAccess     from '../components/CheckAccess.jsx'

const TABS = [
  { key: 'envs',    label: 'สภาพแวดล้อม', icon: '🗂️' },
  { key: 'apps',    label: 'แอปพลิเคชัน', icon: '📦' },
  { key: 'ips',     label: 'IP ที่อนุญาต', icon: '🌐' },
  { key: 'users',   label: 'ผู้ใช้ AD',     icon: '👤' },
  { key: 'check',   label: 'ทดสอบสิทธิ์',  icon: '✅' },
  { key: 'audit',   label: 'Audit Log',    icon: '📜' },
]

export default function Dashboard({ user, onLogout }) {
  const [tab, setTab] = useState('envs')

  return (
    <div className="min-h-full flex">
      {/* Sidebar */}
      <aside className="w-60 shrink-0 bg-slate-900 text-slate-100 flex flex-col">
        <div className="px-5 py-5 border-b border-slate-700">
          <div className="flex items-center gap-2">
            <span className="text-xl">🔐</span>
            <div>
              <div className="font-semibold">SSO Login</div>
              <div className="text-xs text-slate-400">Permission Controller</div>
            </div>
          </div>
        </div>
        <nav className="flex-1 px-2 py-3 space-y-1">
          {TABS.map(t => (
            <button
              key={t.key}
              onClick={() => setTab(t.key)}
              className={
                'w-full text-left px-3 py-2 rounded-md text-sm flex items-center gap-2 transition-colors ' +
                (tab === t.key
                  ? 'bg-brand-600 text-white'
                  : 'text-slate-300 hover:bg-slate-800')
              }
            >
              <span>{t.icon}</span><span>{t.label}</span>
            </button>
          ))}
        </nav>
        <div className="px-4 py-3 border-t border-slate-700 text-sm">
          <div className="text-slate-300">
            ผู้ใช้: <span className="font-medium">{user?.displayName || user?.username}</span>
          </div>
          <button onClick={onLogout} className="mt-2 w-full text-left text-rose-300 hover:text-rose-200">
            ↩ ออกจากระบบ
          </button>
        </div>
      </aside>

      {/* Main */}
      <main className="flex-1 overflow-y-auto">
        <header className="bg-white border-b border-slate-200 px-6 py-3 flex items-center justify-between">
          <h1 className="text-lg font-semibold text-slate-800">
            {TABS.find(t => t.key === tab)?.label}
          </h1>
          <div className="text-sm text-slate-500">
            session expires: {user?.expiresAt ? new Date(user.expiresAt).toLocaleString('th-TH') : '-'}
          </div>
        </header>
        <div className="p-6">
          {tab === 'envs'  && <EnvManager />}
          {tab === 'apps'  && <AppManager />}
          {tab === 'ips'   && <AllowedIpManager />}
          {tab === 'users' && <AllowedUserManager />}
          {tab === 'check' && <CheckAccess />}
          {tab === 'audit' && <AuditLog />}
        </div>
      </main>
    </div>
  )
}
