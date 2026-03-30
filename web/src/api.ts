const apiHeaders = (token: string): Record<string, string> => ({
  'Content-Type': 'application/json',
  'X-Api-Token': token,
})

async function extractError(res: Response): Promise<string> {
  const data = await res.json().catch(() => ({}))
  return (data as { error?: string }).error ?? `HTTP ${res.status}`
}

export interface ApiToken {
  token: string
  admin: boolean
}

export interface TokenPermissions {
  maintainer: boolean
  download: boolean
  upload: boolean
  delete: boolean
}

export interface Version {
  path: string
  checksum: string
}

export interface Product {
  name: string
  group_name: string
  tokens: Record<string, TokenPermissions>
  versions: Record<string, Version>
}

export interface HealthStatus {
  mysql: string
  redis: string
}

export async function validateLogin(token: string): Promise<boolean> {
  const res = await fetch('/ui/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token }),
  })
  if (!res.ok) throw new Error(await extractError(res))
  const data = await res.json()
  return (data as { admin?: boolean }).admin === true
}

export async function listTokens(token: string): Promise<ApiToken[]> {
  const res = await fetch('/api/tokens', { headers: apiHeaders(token) })
  if (!res.ok) throw new Error(await extractError(res))
  return res.json()
}

export async function createToken(token: string, admin: boolean): Promise<string> {
  const res = await fetch('/api/register', {
    method: 'PUT',
    headers: apiHeaders(token),
    body: JSON.stringify({ admin }),
  })
  if (!res.ok) throw new Error(await extractError(res))
  return res.text()
}

export async function revokeToken(token: string, targetToken: string): Promise<void> {
  const res = await fetch(`/api/prune/${encodeURIComponent(targetToken)}`, {
    method: 'DELETE',
    headers: apiHeaders(token),
  })
  if (!res.ok) throw new Error(await extractError(res))
}

export async function listProducts(token: string): Promise<Product[]> {
  const res = await fetch('/api/product/list', { headers: apiHeaders(token) })
  if (!res.ok) throw new Error(await extractError(res))
  return res.json()
}

export async function fetchProduct(token: string, name: string, group: string): Promise<Product> {
  const params = group ? `?group=${encodeURIComponent(group)}` : ''
  const res = await fetch(`/api/product/fetch/${encodeURIComponent(name)}${params}`, {
    headers: apiHeaders(token),
  })
  if (!res.ok) throw new Error(await extractError(res))
  return res.json()
}

export async function createProduct(token: string, name: string, group: string): Promise<void> {
  const res = await fetch('/api/product/create', {
    method: 'PUT',
    headers: apiHeaders(token),
    body: JSON.stringify({ name, group_name: group, tokens: {} }),
  })
  if (!res.ok) throw new Error(await extractError(res))
}

export async function deleteProduct(token: string, name: string, group: string): Promise<void> {
  const params = group ? `?group=${encodeURIComponent(group)}` : ''
  const res = await fetch(`/api/product/delete/${encodeURIComponent(name)}${params}`, {
    method: 'DELETE',
    headers: apiHeaders(token),
  })
  if (!res.ok) throw new Error(await extractError(res))
}

export async function downloadVersion(
  token: string,
  product: string,
  group: string,
  version: string,
): Promise<void> {
  const params = group ? `?group=${encodeURIComponent(group)}` : ''
  const res = await fetch(
    `/api/product/download/${encodeURIComponent(product)}/${encodeURIComponent(version)}${params}`,
    { headers: { 'X-Api-Token': token } },
  )
  if (!res.ok) throw new Error(await extractError(res))

  const blob = await res.blob()

  // Derive filename from Content-Disposition, fall back to product-version
  const disposition = res.headers.get('Content-Disposition') ?? ''
  const match = disposition.match(/filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/)
  const filename = match ? match[1].replace(/['"]/g, '') : `${product}-${version}`

  const url = URL.createObjectURL(blob)
  const a   = document.createElement('a')
  a.href     = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

export async function uploadVersion(
  token: string,
  product: string,
  group: string,
  version: string,
  file: File,
): Promise<void> {
  const form = new FormData()
  form.append('product', product)
  if (group) form.append('group_name', group)
  form.append('version', version)
  form.append('file', file)
  // Do not set Content-Type — fetch sets it with the multipart boundary automatically
  const res = await fetch('/api/product/upload', {
    method: 'POST',
    headers: { 'X-Api-Token': token },
    body: form,
  })
  if (!res.ok) throw new Error(await extractError(res))
}

export async function deleteVersion(
  token: string,
  product: string,
  group: string,
  version: string,
): Promise<void> {
  const params = group ? `?group=${encodeURIComponent(group)}` : ''
  const res = await fetch(
    `/api/product/delete/${encodeURIComponent(product)}/${encodeURIComponent(version)}${params}`,
    { method: 'DELETE', headers: apiHeaders(token) },
  )
  if (!res.ok) throw new Error(await extractError(res))
}

export async function fetchHealth(token: string): Promise<HealthStatus> {
  const res = await fetch('/api/health', { headers: apiHeaders(token) })
  const data: HealthStatus = await res.json().catch(() => ({
    mysql: 'unreachable',
    redis: 'unreachable',
  }))
  return data
}
