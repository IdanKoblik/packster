import { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { api, Product, ApiToken, ApiError } from '../api/client';
import './Dashboard.css';

export function Dashboard() {
  const { token, logout } = useAuth();
  const [activeTab, setActiveTab] = useState<'products' | 'tokens'>('products');
  const [products, setProducts] = useState<Product[]>([]);
  const [tokens, setTokens] = useState<{ token: string; admin: boolean }[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [newProductName, setNewProductName] = useState('');
  const [newTokenAdmin, setNewTokenAdmin] = useState(false);
  const [showNewToken, setShowNewToken] = useState<string | null>(null);
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    if (!token) return;
    setLoading(true);
    setError(null);
    try {
      const health = await api.getHealth(token);
      console.log('Health:', health);
      await loadProducts();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to load data');
    } finally {
      setLoading(false);
    }
  };

  const loadProducts = async () => {
    if (!token) return;
    try {
      const response = await fetch('/api/product/fetch', {
        headers: { 'X-Api-Token': token },
      });
      if (response.ok) {
        const data = await response.json();
        setProducts(Array.isArray(data) ? data : []);
      }
    } catch (err) {
      console.error('Failed to fetch products:', err);
    }
  };

  const handleCreateProduct = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!token || !newProductName.trim()) return;

    try {
      await api.createProduct(token, newProductName.trim());
      setNewProductName('');
      await loadProducts();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to create product');
    }
  };

  const handleDeleteProduct = async (productName: string) => {
    if (!token || !confirm(`Delete product "${productName}"?`)) return;

    try {
      await api.deleteProduct(token, productName);
      await loadProducts();
      if (selectedProduct?.name === productName) {
        setSelectedProduct(null);
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to delete product');
    }
  };

  const handleRegisterToken = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!token) return;

    try {
      const newTokenValue = await api.registerToken(token, newTokenAdmin);
      setTokens([...tokens, { token: newTokenValue, admin: newTokenAdmin }]);
      setShowNewToken(newTokenValue);
      setNewTokenAdmin(false);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to register token');
    }
  };

  const handlePruneToken = async (tokenToPrune: string) => {
    if (!token || !confirm('Delete this token? This action cannot be undone.')) return;

    try {
      await api.pruneToken(token, tokenToPrune);
      setTokens(tokens.filter((t) => t.token !== tokenToPrune));
      setShowNewToken(null);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to delete token');
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  if (loading) {
    return <div className="loading">Loading...</div>;
  }

  return (
    <div className="dashboard">
      <header className="dashboard-header">
        <h1>Artifactor Admin</h1>
        <button onClick={logout} className="logout-button">
          Logout
        </button>
      </header>

      {error && (
        <div className="error-banner">
          {error}
          <button onClick={() => setError(null)} className="dismiss-button">
            Dismiss
          </button>
        </div>
      )}

      <nav className="tabs">
        <button
          className={`tab ${activeTab === 'products' ? 'active' : ''}`}
          onClick={() => setActiveTab('products')}
        >
          Products
        </button>
        <button
          className={`tab ${activeTab === 'tokens' ? 'active' : ''}`}
          onClick={() => setActiveTab('tokens')}
        >
          Tokens
        </button>
      </nav>

      <main className="dashboard-content">
        {activeTab === 'products' ? (
          <div className="products-section">
            <form onSubmit={handleCreateProduct} className="create-form">
              <input
                type="text"
                value={newProductName}
                onChange={(e) => setNewProductName(e.target.value)}
                placeholder="New product name"
              />
              <button type="submit" disabled={!newProductName.trim()}>
                Create Product
              </button>
            </form>

            <div className="products-grid">
              {products.length === 0 ? (
                <div className="empty-state">No products yet</div>
              ) : (
                products.map((product) => (
                  <div
                    key={product.name}
                    className={`product-card ${selectedProduct?.name === product.name ? 'selected' : ''}`}
                    onClick={() => setSelectedProduct(product)}
                  >
                    <div className="product-header">
                      <h3>{product.name}</h3>
                      <button
                        className="delete-icon"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleDeleteProduct(product.name);
                        }}
                      >
                        Delete
                      </button>
                    </div>
                    <div className="product-stats">
                      <span>{Object.keys(product.versions || {}).length} versions</span>
                      <span>{Object.keys(product.tokens || {}).length} tokens</span>
                    </div>
                  </div>
                ))
              )}
            </div>

            {selectedProduct && (
              <div className="product-detail">
                <h2>{selectedProduct.name}</h2>
                <div className="detail-section">
                  <h3>Versions ({Object.keys(selectedProduct.versions || {}).length})</h3>
                  {Object.keys(selectedProduct.versions || {}).length === 0 ? (
                    <p className="empty-message">No versions</p>
                  ) : (
                    <ul className="version-list">
                      {Object.entries(selectedProduct.versions || {}).map(([version, data]) => (
                        <li key={version}>
                          <span className="version-name">{version}</span>
                          <span className="version-checksum">{data.checksum}</span>
                        </li>
                      ))}
                    </ul>
                  )}
                </div>
                <div className="detail-section">
                  <h3>Tokens ({Object.keys(selectedProduct.tokens || {}).length})</h3>
                  {Object.keys(selectedProduct.tokens || {}).length === 0 ? (
                    <p className="empty-message">No tokens with access</p>
                  ) : (
                    <ul className="token-list">
                      {Object.entries(selectedProduct.tokens || {}).map(([hashedToken, perms]) => (
                        <li key={hashedToken}>
                          <span className="token-hash">{hashedToken.slice(0, 12)}...</span>
                          <span className="token-perms">
                            {perms.maintainer && 'M'}
                            {perms.upload && 'U'}
                            {perms.download && 'D'}
                            {perms.delete && 'X'}
                          </span>
                        </li>
                      ))}
                    </ul>
                  )}
                </div>
                <button onClick={() => setSelectedProduct(null)} className="close-detail">
                  Close
                </button>
              </div>
            )}
          </div>
        ) : (
          <div className="tokens-section">
            <form onSubmit={handleRegisterToken} className="create-form">
              <label className="checkbox-label">
                <input
                  type="checkbox"
                  checked={newTokenAdmin}
                  onChange={(e) => setNewTokenAdmin(e.target.checked)}
                />
                Admin privileges
              </label>
              <button type="submit">Register Token</button>
            </form>

            {showNewToken && (
              <div className="new-token-banner">
                <p>New token created (admin: {newTokenAdmin ? 'yes' : 'no'}):</p>
                <code>{showNewToken}</code>
                <div className="new-token-actions">
                  <button onClick={() => copyToClipboard(showNewToken)}>Copy</button>
                  <button onClick={() => setShowNewToken(null)}>Dismiss</button>
                </div>
              </div>
            )}

            <div className="tokens-list">
              <h3>Registered Tokens</h3>
              {tokens.length === 0 ? (
                <p className="empty-message">No tokens registered yet</p>
              ) : (
                <table className="tokens-table">
                  <thead>
                    <tr>
                      <th>Token</th>
                      <th>Type</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {tokens.map((t) => (
                      <tr key={t.token}>
                        <td className="token-value">{t.token.slice(0, 20)}...</td>
                        <td>{t.admin ? 'Admin' : 'Regular'}</td>
                        <td>
                          <button
                            onClick={() => copyToClipboard(t.token)}
                            className="action-button"
                          >
                            Copy
                          </button>
                          <button
                            onClick={() => handlePruneToken(t.token)}
                            className="action-button delete"
                          >
                            Delete
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          </div>
        )}
      </main>
    </div>
  );
}
