import { useState, useEffect, useCallback } from 'react'
import { fetchHealth, HealthStatus } from '../api'

interface Props {
  token: string
}

export default function Health({ token }: Props) {
  const [status, setStatus]   = useState<HealthStatus | null>(null)
  const [loading, setLoading] = useState(true)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      setStatus(await fetchHealth(token))
    } catch {
      setStatus({ mysql: 'unreachable', redis: 'unreachable' })
    } finally {
      setLoading(false)
    }
  }, [token])

  useEffect(() => { load() }, [load])

  const isOk = (s: string) => s.toLowerCase().includes('fine')

  const services: { label: string; value: string }[] = status
    ? [
        { label: 'MySQL', value: status.mysql },
        { label: 'Redis', value: status.redis },
      ]
    : []

  return (
    <>
      <div className="section-header">
        <span className="section-title">System Health</span>
        <button className="btn btn-secondary btn-sm" onClick={load}>
          Refresh
        </button>
      </div>

      {loading ? (
        <div className="loading">Checking…</div>
      ) : (
        <div className="health-grid">
          {services.map(({ label, value }) => {
            const ok = isOk(value)
            return (
              <div key={label} className="health-card">
                <div className="health-service">{label}</div>
                <div className={`health-status ${ok ? 'col-ok' : 'col-err'}`}>
                  <span className={`dot ${ok ? 'dot-ok' : 'dot-err'}`} />
                  {value}
                </div>
              </div>
            )
          })}
        </div>
      )}
    </>
  )
}
