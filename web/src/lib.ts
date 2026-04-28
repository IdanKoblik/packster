import type { Host } from "./types";

export const MOCK_HOSTS: Host[] = [
  { id: 1, type: "gitlab", url: "https://gitlab.com" },
  { id: 2, type: "github", url: "https://github.com" },
  { id: 3, type: "gitlab", url: "https://git.internal.packster.dev" },
  { id: 4, type: "github", url: "http://ghe.staging.local" },
];

export interface ParsedHost {
  scheme: string;
  host: string;
  secure: boolean;
}

export function parseHost(rawUrl: string): ParsedHost {
  try {
    const u = new URL(rawUrl);
    return {
      scheme: u.protocol.replace(":", ""),
      host: u.host,
      secure: u.protocol === "https:",
    };
  } catch {
    return { scheme: "", host: rawUrl, secure: false };
  }
}

export function providerLabel(type: Host["type"]): string {
  if (type === "github") return "GitHub";
  if (type === "gitlab") return "GitLab";
  return type;
}

export function isLocalhost(): boolean {
  const h = window.location.hostname;
  return (
    h === "localhost" ||
    h === "127.0.0.1" ||
    h === "::1" ||
    window.location.protocol === "file:"
  );
}
