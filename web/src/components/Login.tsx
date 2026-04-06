import { useState } from 'react'

interface Props {
  onLogin: (token: string) => void
}

const features = [
  { icon: '📦', label: 'Package Management', desc: 'Organize and track all your product versions in one place' },
  { icon: '🔗', label: 'Git Integration', desc: 'Seamlessly connect releases to your Git repositories' },
  { icon: '🔒', label: 'Secure Access', desc: 'Role-based permissions with admin and regular user tiers' },
]

export default function Login({ onLogin }: Props) {
  const [error, setError] = useState('')
  const [gitlabLoading, setGitlabLoading] = useState(false)

  return (
    <div className="login-wrapper">
      <div className="login-card">
        <div className="login-header">
          <div className="login-logo-text">PACKSTER</div>
          <p className="login-tagline">Product Version Manager</p>
        </div>

        <div className="login-features">
          {features.map((f) => (
            <div key={f.label} className="login-feature-item">
              <span className="login-feature-icon">{f.icon}</span>
              <div>
                <div className="login-feature-label">{f.label}</div>
                <div className="login-feature-desc">{f.desc}</div>
              </div>
            </div>
          ))}
        </div>

        <div className="login-divider">Sign in to continue</div>

       {error && <div className="alert alert-error">{error}</div>}

       <button
          type="button"
          className="btn btn-gitlab btn-full"
          disabled={gitlabLoading}
          onClick={async () => {
            setError('')
            setGitlabLoading(true)
            try {
              const res = await fetch('/api/auth/gitlab/status')
              if (!res.ok) throw new Error('GitLab login is not enabled')
              window.location.href = '/api/auth/gitlab/redirect'
            } catch (err: unknown) {
              setError(err instanceof Error ? err.message : 'GitLab login is not enabled')
            } finally {
              setGitlabLoading(false)
            }
          }}
        >
        <svg className="gitlab-icon" viewBox="0 0 24 24" aria-hidden="true">
            <path d="M22.65 14.39L12 22.13 1.35 14.39a.84.84 0 0 1-.3-.94l1.22-3.78 2.44-7.51A.42.42 0 0 1 4.82 2a.43.43 0 0 1 .58 0 .42.42 0 0 1 .11.18l2.44 7.49h8.1l2.44-7.51A.42.42 0 0 1 18.6 2a.43.43 0 0 1 .58 0 .42.42 0 0 1 .11.18l2.44 7.51 1.22 3.78a.84.84 0 0 1-.3.92z"/>
        </svg>
        {gitlabLoading ? 'Checking…' : 'Sign in with GitLab'}
        </button>

      </div>
    </div>
  )
}
