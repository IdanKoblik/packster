import { useState } from 'react'
import Tokens from './Tokens'
import Products from './Products'
import Health from './Health'

type Tab = 'tokens' | 'products' | 'health'

interface Props {
  token: string
  isAdmin: boolean
  onLogout: () => void
}

const ALL_TABS: { id: Tab; label: string; adminOnly?: boolean }[] = [
  { id: 'tokens',   label: 'Tokens',   adminOnly: true },
  { id: 'products', label: 'Products' },
  { id: 'health',   label: 'Health'   },
]

export default function Dashboard({ token, isAdmin, onLogout }: Props) {
  const tabs = ALL_TABS.filter(t => !t.adminOnly || isAdmin)
  const [tab, setTab] = useState<Tab>(isAdmin ? 'tokens' : 'products')
  const [confirmLogout, setConfirmLogout] = useState(false)

  return (
    <div className="dashboard">
      <header className="header">
        <div className="header-inner">
          <span className="header-title">PACKSTER</span>
          <div className="header-right">
            {isAdmin && <span className="badge-admin">admin</span>}
            <button className="btn btn-secondary btn-sm" onClick={() => setConfirmLogout(true)}>
              Logout
            </button>
          </div>
        </div>
      </header>

      <nav className="nav">
        <div className="nav-inner">
          {tabs.map(t => (
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
        {tab === 'products' && <Products token={token} isAdmin={isAdmin} />}
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
