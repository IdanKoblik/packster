import { useState } from 'react'
import Login from './components/Login'

const JWT_TOKEN = 'token'
const TOKEN_TYPE = 'type'

export default function App() {
  const [token, setToken] = useState<string | null>(
    () => sessionStorage.getItem(JWT_TOKEN),
  )

  const handleLogin = (t: string) => {
    sessionStorage.setItem(JWT_TOKEN, t)
    sessionStorage.setItem(TOKEN_TYPE, "gitlab")
    setToken(t)
  }

  if (!token) return <Login onLogin={handleLogin} />
  return <h1>TODO</h1>
}
