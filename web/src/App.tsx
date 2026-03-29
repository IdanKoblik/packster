import { useState } from 'react'
import Login from './components/Login'
import Dashboard from './components/Dashboard'

const TOKEN_KEY = 'packster_token'
const ADMIN_KEY = 'packster_admin'

export default function App() {
  const [token, setToken] = useState<string | null>(
    () => sessionStorage.getItem(TOKEN_KEY),
  )
  const [isAdmin, setIsAdmin] = useState<boolean>(
    () => sessionStorage.getItem(ADMIN_KEY) === 'true',
  )

  const handleLogin = (t: string, admin: boolean) => {
    sessionStorage.setItem(TOKEN_KEY, t)
    sessionStorage.setItem(ADMIN_KEY, String(admin))
    setToken(t)
    setIsAdmin(admin)
  }

  const handleLogout = () => {
    sessionStorage.removeItem(TOKEN_KEY)
    sessionStorage.removeItem(ADMIN_KEY)
    setToken(null)
    setIsAdmin(false)
  }

  if (!token) return <Login onLogin={handleLogin} />
  return <Dashboard token={token} isAdmin={isAdmin} onLogout={handleLogout} />
}
