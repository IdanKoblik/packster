export type HostType = "github" | "gitlab";

export interface Host {
  id: number;
  type: HostType;
  url: string;
}

export type LoadState =
  | { kind: "loading" }
  | { kind: "success"; hosts: Host[] }
  | { kind: "empty" }
  | { kind: "error" };

export type HealthState = "pending" | "ok" | "down";

export type DebugState = "success" | "loading" | "empty" | "error";

export interface Jwt {
  token: string;
  host: { type: HostType; url: string };
  orgs: number[];
  sub?: string;
  name?: string;
  exp?: number;
}

export interface Project {
  id: number;
  name: string;
  org: number;
  web_url: string;
  repository?: number;
  owner?: number;
}

export interface Product {
  id: number;
  name: string;
  project: number;
}

export interface Version {
  id: number;
  name: string;
  path: string;
  checksum: string;
  product: number;
}

export interface Permission {
  account: number;
  project: number;
  can_download: boolean;
  can_upload: boolean;
  can_delete: boolean;
}

export interface PermissionEntry {
  user_id: number;
  display_name: string;
  project: number;
  can_download: boolean;
  can_upload: boolean;
  can_delete: boolean;
  is_owner: boolean;
}

export interface UserCandidate {
  id: number;
  display_name: string;
}

export type ProductsState =
  | { kind: "loading" }
  | { kind: "success"; products: Product[] }
  | { kind: "empty" }
  | { kind: "error"; message?: string };

export type VersionsState =
  | { kind: "loading" }
  | { kind: "success"; versions: Version[] }
  | { kind: "empty" }
  | { kind: "error"; message?: string };

export interface CandidateProject {
  id: number;
  name: string;
  full_path: string;
  web_url: string;
  org: number;
  visibility?: "public" | "private" | "internal";
  last_activity_at?: string;
}

export type CandidatesState =
  | { kind: "loading" }
  | { kind: "success"; candidates: CandidateProject[] }
  | { kind: "empty" }
  | { kind: "error"; message?: string };

export type ProjectsState =
  | { kind: "loading" }
  | { kind: "success"; projects: Project[] }
  | { kind: "empty" }
  | { kind: "error"; message?: string };

export interface Toast {
  id: number;
  kind: "info" | "error" | "success";
  message: string;
}
