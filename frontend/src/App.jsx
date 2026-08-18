import { useState, useEffect } from 'react'
import Login from './pages/Login.jsx'
import Dashboard from './pages/Dashboard.jsx'
import { getToken, getUser, clearAuth } from './api.js'

export default function App() {
  const [authed, setAuthed] = useState(!!getToken())
  const [user, setUser] = useState(getUser())

  useEffect(() => {
    const onUnauth = () => { setAuthed(false); setUser(null) }
    window.addEventListener('sso:unauthorized', onUnauth)
    return () => window.removeEventListener('sso:unauthorized', onUnauth)
  }, [])

  const handleLogin = (u) => {
    setAuthed(true)
    setUser(u)
  }

  const handleLogout = () => {
    clearAuth()
    setAuthed(false)
    setUser(null)
  }

  if (!authed) return <Login onLogin={handleLogin} />
  return <Dashboard user={user} onLogout={handleLogout} />
}
