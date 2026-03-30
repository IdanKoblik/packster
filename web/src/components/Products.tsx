import { useState, useEffect, useCallback, useRef } from 'react'
import {
  listProducts,
  fetchProduct,
  createProduct,
  deleteProduct,
  uploadVersion,
  deleteVersion,
  downloadVersion,
  Product,
} from '../api'

interface Props {
  token: string
  isAdmin: boolean
}

export default function Products({ token, isAdmin }: Props) {
  const [products, setProducts]   = useState<Product[] | null>(null)
  const [loadError, setLoadError] = useState('')
  const [loading, setLoading]     = useState(true)

  const [selected, setSelected]           = useState<Product | null>(null)
  const [detailLoading, setDetailLoading] = useState(false)
  const [detailError, setDetailError]     = useState('')

  const [showCreateModal, setShowCreateModal] = useState(false)
  const [newName, setNewName]                 = useState('')
  const [newGroup, setNewGroup]               = useState('')
  const [createError, setCreateError]         = useState('')

  const [downloadingVer, setDownloadingVer] = useState<string | null>(null)

  const [showUploadModal, setShowUploadModal] = useState(false)
  const [uploadVer, setUploadVer]             = useState('')
  const [uploadFile, setUploadFile]           = useState<File | null>(null)
  const [uploadError, setUploadError]         = useState('')
  const [uploading, setUploading]             = useState(false)
  const fileInputRef                          = useRef<HTMLInputElement>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setLoadError('')
    try {
      setProducts(await listProducts(token))
    } catch (e: unknown) {
      setLoadError(e instanceof Error ? e.message : 'Failed to load products')
    } finally {
      setLoading(false)
    }
  }, [token])

  useEffect(() => { load() }, [load])

  const refreshSelected = async (name: string, group: string) => {
    try {
      setSelected(await fetchProduct(token, name, group))
    } catch {
      // ignore — stale detail is acceptable
    }
  }

  const handleView = async (name: string, group: string) => {
    setDetailLoading(true)
    setDetailError('')
    setSelected(null)
    try {
      setSelected(await fetchProduct(token, name, group))
    } catch (e: unknown) {
      setDetailError(e instanceof Error ? e.message : 'Failed to load product')
    } finally {
      setDetailLoading(false)
    }
  }

  const handleDelete = async (name: string, group: string) => {
    const label = group ? `${group}/${name}` : name
    if (!confirm(`Delete product "${label}"?\nThis cannot be undone.`)) return
    setLoadError('')
    try {
      await deleteProduct(token, name, group)
      if (selected?.name === name && selected?.group_name === group) setSelected(null)
      load()
    } catch (e: unknown) {
      setLoadError(e instanceof Error ? e.message : 'Failed to delete product')
    }
  }

  const openCreateModal = () => {
    setNewName('')
    setNewGroup('')
    setCreateError('')
    setShowCreateModal(true)
  }

  const handleCreate = async () => {
    const name = newName.trim()
    if (!name) { setCreateError('Product name is required'); return }
    setCreateError('')
    try {
      await createProduct(token, name, newGroup.trim())
      setShowCreateModal(false)
      load()
    } catch (e: unknown) {
      setCreateError(e instanceof Error ? e.message : 'Failed to create product')
    }
  }

  const openUploadModal = () => {
    setUploadVer('')
    setUploadFile(null)
    setUploadError('')
    setUploading(false)
    if (fileInputRef.current) fileInputRef.current.value = ''
    setShowUploadModal(true)
  }

  const handleUpload = async () => {
    if (!selected) return
    const ver = uploadVer.trim()
    if (!ver)        { setUploadError('Version name is required'); return }
    if (!uploadFile) { setUploadError('File is required');         return }

    setUploadError('')
    setUploading(true)
    try {
      await uploadVersion(token, selected.name, selected.group_name, ver, uploadFile)
      setShowUploadModal(false)
      refreshSelected(selected.name, selected.group_name)
    } catch (e: unknown) {
      setUploadError(e instanceof Error ? e.message : 'Upload failed')
    } finally {
      setUploading(false)
    }
  }

  const handleDownloadVersion = async (ver: string) => {
    if (!selected || downloadingVer) return
    setDownloadingVer(ver)
    try {
      await downloadVersion(token, selected.name, selected.group_name, ver)
    } catch (e: unknown) {
      setDetailError(e instanceof Error ? e.message : 'Download failed')
    } finally {
      setDownloadingVer(null)
    }
  }

  const handleDeleteVersion = async (ver: string) => {
    if (!selected) return
    if (!confirm(`Delete version "${ver}" from "${selected.name}"?\nThis cannot be undone.`)) return
    try {
      await deleteVersion(token, selected.name, selected.group_name, ver)
      refreshSelected(selected.name, selected.group_name)
    } catch (e: unknown) {
      setDetailError(e instanceof Error ? e.message : 'Failed to delete version')
    }
  }

  return (
    <>
      <div className="section-header">
        <span className="section-title">Products</span>
        <button className="btn btn-primary btn-sm" onClick={openCreateModal}>
          + Create Product
        </button>
      </div>

      {loadError && <div className="alert alert-error">{loadError}</div>}

      {loading ? (
        <div className="loading">Loading…</div>
      ) : products && products.length > 0 ? (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Name</th>
                <th>Group</th>
                <th style={{ width: 160 }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {products.map(p => (
                <tr key={`${p.group_name}/${p.name}`}>
                  <td>{p.name}</td>
                  <td>{p.group_name || <span style={{ opacity: 0.4 }}>—</span>}</td>
                  <td>
                    <div className="flex-gap">
                      <button
                        className="btn btn-secondary btn-sm"
                        onClick={() => handleView(p.name, p.group_name)}
                      >
                        View
                      </button>
                      <button
                        className="btn btn-danger btn-sm"
                        onClick={() => handleDelete(p.name, p.group_name)}
                      >
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="empty">No products created yet.</div>
      )}

      {(detailLoading || selected || detailError) && (
        <div className="detail-panel">
          {detailLoading && <div className="loading">Loading…</div>}
          {detailError   && <div className="alert alert-error">{detailError}</div>}
          {selected && (
            <>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 14 }}>
                <span style={{ fontSize: 14 }}>
                  {selected.group_name ? `${selected.group_name} / ` : ''}{selected.name}
                </span>
                <div className="flex-gap">
                  <span className="code">
                    {Object.keys(selected.tokens  ?? {}).length} token(s) ·{' '}
                    {Object.keys(selected.versions ?? {}).length} version(s)
                  </span>
                  <button className="btn btn-primary btn-sm" onClick={openUploadModal}>
                    + Upload Version
                  </button>
                </div>
              </div>

              {Object.keys(selected.versions ?? {}).length > 0 ? (
                <div className="table-wrap" style={{ marginBottom: 0 }}>
                  <table>
                    <thead>
                      <tr>
                        <th>Version</th>
                        <th>Checksum</th>
                        <th style={{ width: 160 }}></th>
                      </tr>
                    </thead>
                    <tbody>
                      {Object.entries(selected.versions).map(([ver, info]) => (
                        <tr key={ver}>
                          <td>{ver}</td>
                          <td className="code">
                            {info.checksum ? `${info.checksum.substring(0, 20)}…` : '—'}
                          </td>
                          <td>
                            <div className="flex-gap">
                              <button
                                className="btn btn-secondary btn-sm"
                                onClick={() => handleDownloadVersion(ver)}
                                disabled={downloadingVer === ver}
                              >
                                {downloadingVer === ver ? '…' : 'Download'}
                              </button>
                              <button
                                className="btn btn-danger btn-sm"
                                onClick={() => handleDeleteVersion(ver)}
                              >
                                Delete
                              </button>
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <div className="empty" style={{ padding: 16 }}>
                  No versions uploaded yet.
                </div>
              )}
            </>
          )}
        </div>
      )}

      {/* ── Create product modal ── */}
      {showCreateModal && (
        <div
          className="modal-overlay"
          onClick={e => { if (e.target === e.currentTarget) setShowCreateModal(false) }}
        >
          <div className="modal">
            <div className="modal-title">Create New Product</div>

            <div className="form-group">
              <label htmlFor="product-name">Product Name</label>
              <input
                id="product-name"
                type="text"
                value={newName}
                onChange={e => setNewName(e.target.value)}
                placeholder="e.g. my-service"
                autoFocus
                onKeyDown={e => e.key === 'Enter' && handleCreate()}
              />
            </div>

            <div className="form-group">
              <label htmlFor="product-group">Group <span style={{ opacity: 0.5, fontWeight: 400 }}>(optional)</span></label>
              <input
                id="product-group"
                type="text"
                value={newGroup}
                onChange={e => setNewGroup(e.target.value)}
                placeholder="e.g. staging"
                onKeyDown={e => e.key === 'Enter' && handleCreate()}
              />
            </div>

            {createError && <div className="alert alert-error">{createError}</div>}

            <div className="modal-footer">
              <button className="btn btn-secondary" onClick={() => setShowCreateModal(false)}>
                Cancel
              </button>
              <button className="btn btn-primary" onClick={handleCreate}>
                Create
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── Upload version modal ── */}
      {showUploadModal && selected && (
        <div
          className="modal-overlay"
          onClick={e => { if (e.target === e.currentTarget) setShowUploadModal(false) }}
        >
          <div className="modal">
            <div className="modal-title">
              Upload Version — {selected.group_name ? `${selected.group_name} / ` : ''}{selected.name}
            </div>

            <div className="form-group">
              <label htmlFor="upload-ver">Version</label>
              <input
                id="upload-ver"
                type="text"
                value={uploadVer}
                onChange={e => setUploadVer(e.target.value)}
                placeholder="e.g. 1.0.0"
                autoFocus
              />
            </div>

            <div className="form-group">
              <label htmlFor="upload-file">Artifact File</label>
              <input
                id="upload-file"
                type="file"
                ref={fileInputRef}
                onChange={e => setUploadFile(e.target.files?.[0] ?? null)}
              />
            </div>

            {uploadFile && (
              <div className="code" style={{ marginBottom: 14, fontSize: 11 }}>
                {uploadFile.name} · {(uploadFile.size / 1024).toFixed(1)} KB
              </div>
            )}

            {uploadError && <div className="alert alert-error">{uploadError}</div>}

            <div className="modal-footer">
              <button className="btn btn-secondary" onClick={() => setShowUploadModal(false)}>
                Cancel
              </button>
              <button className="btn btn-primary" onClick={handleUpload} disabled={uploading}>
                {uploading ? 'Uploading…' : 'Upload'}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
