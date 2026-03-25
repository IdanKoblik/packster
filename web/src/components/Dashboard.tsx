import { useState } from 'react'
import Tokens from './Tokens'
import Products from './Products'
import Health from './Health'

type Tab = 'tokens' | 'products' | 'health'

interface Props {
  token: string
  onLogout: () => void
}

const TABS: { id: Tab; label: string }[] = [
  { id: 'tokens',   label: 'Tokens'   },
  { id: 'products', label: 'Products' },
  { id: 'health',   label: 'Health'   },
]

export default function Dashboard({ token, onLogout }: Props) {
  const [tab, setTab] = useState<Tab>('tokens')
  const [confirmLogout, setConfirmLogout] = useState(false)

  return (
    <div className="dashboard">
      <header className="header">
        <div className="header-inner">
          <span className="header-title">ARTIFACTOR</span>
          <div className="header-right">
            <span className="badge-admin">admin</span>
            <button className="btn btn-secondary btn-sm" onClick={() => setConfirmLogout(true)}>
              Logout
            </button>
          </div>
        </div>
      </header>

      <nav className="nav">
        <div className="nav-inner">
          {TABS.map(t => (
            <div
              key={t.id}
              className={`nav-tab${tab === t.id ? ' active' : ''}`}
              onClick={() => setTab(t.id)}
            >
              {t.label}
            </div>
          ))}
        </div>
      </nav>

      <main className="content">
        {tab === 'tokens'   && <Tokens   token={token} />}
        {tab === 'products' && <Products token={token} />}
        {tab === 'health'   && <Health   token={token} />}
      </main>

      {confirmLogout && (
        <div className="modal-overlay" onClick={() => setConfirmLogout(false)}>
          <div className="modal" onClick={e => e.stopPropagation()}>
            <div className="modal-title">Confirm Logout</div>
            <p style={{ color: 'var(--muted)', fontSize: '15px' }}>
              Are you sure you want to log out?
            </p>
            <div className="modal-footer">
              <button className="btn btn-secondary" onClick={() => setConfirmLogout(false)}>
                Cancel
              </button>
              <button className="btn btn-danger" onClick={onLogout}>
                Logout
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
