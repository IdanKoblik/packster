import { useState } from 'react'
import Login from './components/Login'
import Dashboard from './components/Dashboard'

const TOKEN_KEY = 'artifactor_token'

function getCookie(name: string): string | null {
  const match = document.cookie.match(new RegExp('(?:^|; )' + name + '=([^;]*)'))
  return match ? decodeURIComponent(match[1]) : null
}

function setCookie(name: string, value: string) {
  document.cookie = `${name}=${encodeURIComponent(value)}; path=/; SameSite=Strict`
}

function deleteCookie(name: string) {
  document.cookie = `${name}=; path=/; max-age=0`
}

export default function App() {
  const [token, setToken] = useState<string | null>(
    () => getCookie(TOKEN_KEY),
  )

  const handleLogin = (t: string) => {
    setCookie(TOKEN_KEY, t)
    setToken(t)
  }

  const handleLogout = () => {
    deleteCookie(TOKEN_KEY)
    setToken(null)
  }

  if (!token) return <Login onLogin={handleLogin} />
  return <Dashboard token={token} onLogout={handleLogout} />
}
