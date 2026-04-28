import type { Permission, PermissionEntry, Product, Project, UserCandidate, Version } from "./types";

export class UnauthorizedError extends Error {
  constructor() {
    super("unauthorized");
  }
}

export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
  }
}

async function request<T>(token: string, path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers ?? {});
  headers.set("Authorization", `Bearer ${token}`);
  if (!headers.has("Accept")) headers.set("Accept", "application/json");

  const res = await fetch(path, { ...init, headers });
  if (res.status === 401 || res.status === 403) throw new UnauthorizedError();

  if (!res.ok) {
    let msg = `${res.status} ${res.statusText}`;
    try {
      const body = await res.json();
      if (body && typeof body.error === "string") msg = body.error;
    } catch {
      /* not JSON */
    }
    throw new ApiError(res.status, msg);
  }

  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

export const listProjects = (token: string) =>
  request<Project[]>(token, "/api/user/projects");

export const deleteProject = (token: string, projectId: number) =>
  request<{ ok: true }>(token, `/api/projects/${projectId}`, { method: "DELETE" });

export const listProducts = (token: string, projectId: number) =>
  request<Product[]>(token, `/api/projects/${projectId}/products`);

export const createProduct = (token: string, projectId: number, name: string) =>
  request<Product>(token, `/api/projects/${projectId}/products`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name }),
  });

export const deleteProduct = (token: string, projectId: number, productId: number) =>
  request<{ ok: true }>(token, `/api/projects/${projectId}/products/${productId}`, {
    method: "DELETE",
  });

export const listVersions = (token: string, productId: number) =>
  request<Version[]>(token, `/api/products/${productId}/versions`);

export async function uploadVersion(
  token: string,
  productId: number,
  name: string,
  file: File,
): Promise<Version> {
  const form = new FormData();
  form.append("name", name);
  form.append("file", file);

  const res = await fetch(`/api/products/${productId}/versions`, {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
    body: form,
  });
  if (res.status === 401 || res.status === 403) throw new UnauthorizedError();
  if (res.status === 413) throw new ApiError(413, "file exceeds upload limit");
  if (!res.ok) {
    let msg = `${res.status} ${res.statusText}`;
    try {
      const body = await res.json();
      if (body && typeof body.error === "string") msg = body.error;
    } catch {
      /* not JSON */
    }
    throw new ApiError(res.status, msg);
  }
  return (await res.json()) as Version;
}

export const deleteVersion = (token: string, versionId: number) =>
  request<{ ok: true }>(token, `/api/versions/${versionId}`, { method: "DELETE" });

// Anchor downloads can't carry an Authorization header, so we fetch the blob
// and trigger a synthetic <a download> instead of a direct link.
export async function downloadVersion(
  token: string,
  versionId: number,
  filename: string,
): Promise<void> {
  const res = await fetch(`/api/versions/${versionId}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (res.status === 401 || res.status === 403) throw new UnauthorizedError();
  if (!res.ok) throw new ApiError(res.status, `download failed: ${res.status}`);
  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}

export const listPermissions = (token: string, projectId: number) =>
  request<PermissionEntry[]>(token, `/api/projects/${projectId}/permissions`);

export const setPermission = (
  token: string,
  projectId: number,
  body: { user_id: number; can_download: boolean; can_upload: boolean; can_delete: boolean },
) =>
  request<Permission>(token, `/api/projects/${projectId}/permissions`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });

export const revokePermission = (token: string, projectId: number, userId: number) =>
  request<{ ok: true }>(token, `/api/projects/${projectId}/permissions/${userId}`, {
    method: "DELETE",
  });

export const searchUserCandidates = (token: string, projectId: number, q: string) =>
  request<UserCandidate[]>(
    token,
    `/api/projects/${projectId}/permissions/candidates?q=${encodeURIComponent(q)}`,
  );
