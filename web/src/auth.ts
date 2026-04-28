import type { Jwt } from "./types";

const KEY = "packster_jwt";

// Decodes the JWT payload without verifying it. Server is the source of truth;
// the client only reads claims for routing/UI.
export function decodeJwt(token: string): Jwt | null {
  try {
    const parts = token.split(".");
    if (parts.length < 2) return null;
    const payload = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    const padded = payload + "=".repeat((4 - (payload.length % 4)) % 4);
    const json = atob(padded);
    const decoded = JSON.parse(json);
    if (!decoded || typeof decoded !== "object") return null;
    if (typeof decoded.token !== "string") return null;
    if (!decoded.host || typeof decoded.host.type !== "string") return null;
    if (typeof decoded.host.url !== "string") return null;
    if (!Array.isArray(decoded.orgs)) return null;
    return decoded as Jwt;
  } catch {
    return null;
  }
}

export function getStoredJwt(): { raw: string; payload: Jwt } | null {
  try {
    const raw = localStorage.getItem(KEY);
    if (!raw) return null;
    const payload = decodeJwt(raw);
    if (!payload) return null;
    if (payload.exp && payload.exp * 1000 < Date.now()) return null;
    return { raw, payload };
  } catch {
    return null;
  }
}

export function setJwt(raw: string) {
  localStorage.setItem(KEY, raw);
}

export function clearJwt() {
  localStorage.removeItem(KEY);
}

// Verifies the JWT against the server so locally-valid tokens for deleted
// accounts don't grant access.
export async function verifySession(token: string): Promise<boolean> {
  try {
    const res = await fetch("/api/auth/session", {
      headers: {
        accept: "application/json",
        authorization: `Bearer ${token}`,
      },
    });
    if (res.status === 401 || res.status === 403) return false;
    return res.ok;
  } catch {
    return true;
  }
}
