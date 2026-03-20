import { useState } from 'react';
import { useAuth } from '../context/AuthContext';
import './Login.css';

export function Login() {
  const { login, isLoading, error } = useAuth();
  const [token, setToken] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [validationError, setValidationError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setValidationError(null);
    setIsSubmitting(true);

    try {
      await login(token);
    } catch (err) {
      setValidationError(err instanceof Error ? err.message : 'Login failed');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="login-container">
      <div className="login-card">
        <h1 className="login-title">Artifactor</h1>
        <p className="login-subtitle">Admin Portal</p>

        <form onSubmit={handleSubmit} className="login-form">
          <div className="form-group">
            <label htmlFor="token">Admin Token</label>
            <input
              id="token"
              type="password"
              value={token}
              onChange={(e) => setToken(e.target.value)}
              placeholder="Enter your admin token"
              disabled={isSubmitting || isLoading}
              autoFocus
            />
          </div>

          {(validationError || error) && (
            <div className="error-message">
              {validationError || error}
            </div>
          )}

          <button
            type="submit"
            className="login-button"
            disabled={!token.trim() || isSubmitting || isLoading}
          >
            {isSubmitting || isLoading ? 'Validating...' : 'Login'}
          </button>
        </form>
      </div>
    </div>
  );
}
