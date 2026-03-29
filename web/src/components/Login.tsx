import { useState, FormEvent } from 'react'
import { validateLogin } from '../api'

interface Props {
  onLogin: (token: string, isAdmin: boolean) => void
}

export default function Login({ onLogin }: Props) {
  const [token, setToken] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    const trimmed = token.trim()
    if (!trimmed) { setError('Token is required'); return }

    setError('')
    setLoading(true)

    try {
      const isAdmin = await validateLogin(trimmed)
      onLogin(trimmed, isAdmin)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-wrapper">
      <div className="login-card">
        <div className="login-logo">
          <div className="login-logo-text">PACKSTER</div>
          <div className="login-subtitle">Admin Interface</div>
        </div>

        <form onSubmit={handleSubmit}>
          {error && <div className="alert alert-error">{error}</div>}

          <div className="form-group">
            <label htmlFor="token-input">Admin Token</label>
            <input
              id="token-input"
              type="password"
              value={token}
              onChange={e => setToken(e.target.value)}
              placeholder="Paste your admin token"
              autoComplete="off"
              autoFocus
            />
          </div>

          <button type="submit" className="btn btn-primary btn-full" disabled={loading}>
            {loading ? 'Signing in…' : 'Sign In'}
          </button>
        </form>
      </div>
    </div>
  )
}
