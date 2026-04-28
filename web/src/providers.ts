import type { CandidateProject, HostType, Project } from "./types";

// GitHub is not implemented yet; calls throw so unfinished paths fail loudly.
export interface ProviderAdapter {
  type: HostType;
  label: string;
  fetchUserProjects(args: { token: string; orgs: number[] }): Promise<Project[]>;
  listCandidates(args: { token: string; org: number }): Promise<CandidateProject[]>;
  importProject(args: { token: string; source_id: number; source_url: string; org: number }): Promise<Project>;
}

const gitlabAdapter: ProviderAdapter = {
  type: "gitlab",
  label: "GitLab",

  async fetchUserProjects({ token, orgs }) {
    const res = await fetch("/api/user/projects", {
      headers: {
        accept: "application/json",
        authorization: `Bearer ${token}`,
      },
    });
    if (!res.ok) throw new Error(`projects fetch failed: ${res.status}`);
    const data = await res.json();
    const list: Project[] = Array.isArray(data) ? data : data.projects ?? [];
    const set = new Set(orgs);
    return list.filter((p) => set.has(p.org));
  },

  async listCandidates({ token, org }) {
    const params = new URLSearchParams({
      org: String(org),
      min_access_level: "50",
    });
    const res = await fetch(`/api/user/candidates?${params.toString()}`, {
      headers: {
        accept: "application/json",
        authorization: `Bearer ${token}`,
      },
    });
    if (!res.ok) throw new Error(`candidates fetch failed: ${res.status}`);
    const data = await res.json();
    return Array.isArray(data) ? data : data.candidates ?? [];
  },

  async importProject({ token, source_id, source_url, org }) {
    const res = await fetch("/api/user/projects", {
      method: "POST",
      headers: {
        accept: "application/json",
        "content-type": "application/json",
        authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ source_id, source_url, org }),
    });
    if (!res.ok) throw new Error(`project import failed: ${res.status}`);
    return res.json();
  },
};

const githubAdapter: ProviderAdapter = {
  type: "github",
  label: "GitHub",
  async fetchUserProjects() {
    throw new Error("GitHub provider is not implemented yet");
  },
  async listCandidates() {
    throw new Error("GitHub provider is not implemented yet");
  },
  async importProject() {
    throw new Error("GitHub provider is not implemented yet");
  },
};

const ADAPTERS: Record<HostType, ProviderAdapter> = {
  gitlab: gitlabAdapter,
  github: githubAdapter,
};

export function getProvider(type: HostType): ProviderAdapter {
  return ADAPTERS[type];
}
