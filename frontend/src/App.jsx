import { useEffect, useMemo, useState } from 'react'
import Login from './pages/Login.jsx'
import Dashboard from './pages/Dashboard.jsx'
import { getToken, getUser, clearAuth } from './api.js'
import { isDashboardTab, loginPath, pathToView, tabToPath } from './navigation.js'

export default function App() {
  const [authed, setAuthed] = useState(!!getToken())
  const [user, setUser] = useState(getUser())
  const [path, setPath] = useState(() => window.location.pathname)

  const view = useMemo(() => pathToView(path), [path])

  useEffect(() => {
    const onUnauth = () => {
      clearAuth()
      setAuthed(false)
      setUser(null)
      navigate(loginPath(), true)
    }
    const onPopState = () => setPath(window.location.pathname)
    window.addEventListener('sso:unauthorized', onUnauth)
    window.addEventListener('popstate', onPopState)
    return () => {
      window.removeEventListener('sso:unauthorized', onUnauth)
      window.removeEventListener('popstate', onPopState)
    }
  }, [])

  useEffect(() => {
    if (!authed && view.type !== 'login') {
      navigate(loginPath(), true)
      return
    }
    if (authed && view.type === 'login') {
      navigate(tabToPath('envs'), true)
    }
  }, [authed, view.type])

  const navigate = (nextPath, replace = false) => {
    const normalized = nextPath || loginPath()
    if (window.location.pathname === normalized) {
      setPath(normalized)
      return
    }
    const action = replace ? 'replaceState' : 'pushState'
    window.history[action]({}, '', normalized)
    setPath(normalized)
  }

  const handleLogin = (u) => {
    setAuthed(true)
    setUser(u)
    navigate(tabToPath('envs'), true)
  }

  const handleLogout = () => {
    clearAuth()
    setAuthed(false)
    setUser(null)
    navigate(loginPath(), true)
  }

  const handleNavigate = (tabKey) => {
    if (isDashboardTab(tabKey)) {
      navigate(tabToPath(tabKey))
    }
  }

  if (!authed) return <Login onLogin={handleLogin} />
  return <Dashboard activeTab={view.tab || 'envs'} user={user} onLogout={handleLogout} onNavigate={handleNavigate} />
}
