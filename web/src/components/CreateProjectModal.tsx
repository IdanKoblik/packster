import { useEffect, useMemo, useRef, useState } from "react";
import type { CandidateProject, CandidatesState, Jwt } from "../types";
import { getProvider } from "../providers";
import { AlertIcon, ArrowIcon, EmptyIcon, GithubIcon, GitlabIcon, LockIcon, RefreshIcon, Spinner } from "./icons";

interface Props {
  jwt: Jwt;
  raw: string;
  importedUrls: Set<string>;
  onClose: () => void;
  onImported: (project: unknown) => void;
  onUnauthorized: () => void;
}

export function CreateProjectModal({ jwt, raw, importedUrls, onClose, onImported, onUnauthorized }: Props) {
  const [org, setOrg] = useState<number | null>(jwt.orgs[0] ?? null);
  const [state, setState] = useState<CandidatesState>({ kind: "loading" });
  const [query, setQuery] = useState("");
  const [selected, setSelected] = useState<CandidateProject | null>(null);
  const [importing, setImporting] = useState(false);
  const [importError, setImportError] = useState<string | null>(null);
  const searchRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  const fetchCandidates = async (nextOrg: number | null) => {
    if (nextOrg == null) {
      setState({ kind: "empty" });
      return;
    }
    setState({ kind: "loading" });
    setSelected(null);
    try {
      const provider = getProvider(jwt.host.type);
      const list = await provider.listCandidates({ token: raw, org: nextOrg });
      if (!list.length) setState({ kind: "empty" });
      else setState({ kind: "success", candidates: list });
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      if (/401|403/.test(msg)) return onUnauthorized();
      setState({ kind: "error", message: msg });
    }
  };

  useEffect(() => {
    fetchCandidates(org);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    searchRef.current?.focus();
  }, [state.kind]);

  const filtered = useMemo(() => {
    if (state.kind !== "success") return [];
    const q = query.trim().toLowerCase();
    if (!q) return state.candidates;
    return state.candidates.filter((c) =>
      c.name.toLowerCase().includes(q) || c.full_path.toLowerCase().includes(q)
    );
  }, [state, query]);

  const doImport = async () => {
    if (!selected || importing) return;
    setImporting(true);
    setImportError(null);
    try {
      const provider = getProvider(jwt.host.type);
      const project = await provider.importProject({
        token: raw,
        source_id: selected.id,
        source_url: selected.web_url,
        org: selected.org,
      });
      onImported(project);
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      if (/401|403/.test(msg)) return onUnauthorized();
      setImportError(msg || "Failed to import project");
      setImporting(false);
    }
  };

  const HostIcon = jwt.host.type === "github" ? GithubIcon : GitlabIcon;
  const hostLabel = jwt.host.type === "github" ? "GitHub" : "GitLab";

  return (
    <div
      className="fixed inset-0 z-40 flex items-center justify-center p-6 bg-black/60 backdrop-blur-sm animate-fade-in"
      onClick={onClose}
      role="dialog"
      aria-modal="true"
      aria-label="Import project"
    >
      <div
        onClick={(e) => e.stopPropagation()}
        className="w-full max-w-[560px] bg-surface border border-border rounded-xl shadow-card flex flex-col max-h-[min(640px,90vh)]"
      >
        {/* Header */}
        <div className="px-6 pt-6 pb-4 border-b border-border">
          <div className="inline-flex items-center gap-2 font-mono text-[11px] tracking-[0.14em] uppercase text-ink-mute mb-3.5">
            <span className="w-1.5 h-1.5 rounded-full bg-accent shadow-[0_0_0_3px_rgba(215,226,122,0.12)]" />
            Import from {hostLabel}
          </div>
          <h2 className="text-lg font-semibold tracking-[-0.01em] text-ink m-0 mb-1">Add a project</h2>
          <p className="text-[13px] text-ink-dim leading-relaxed m-0 mb-4">
            Pick a project from your {hostLabel} orgs to pull into Packster.
          </p>

          {/* Controls: org + search */}
          <div className="flex gap-2">
            <select
              value={org == null ? "" : String(org)}
              onChange={(e) => {
                const v = Number(e.target.value);
                setOrg(v);
                fetchCandidates(v);
              }}
              disabled={state.kind === "loading" || importing || jwt.orgs.length === 0}
              className="bg-[#0e0e12] border border-border rounded-md px-3 py-2 text-[13px] text-ink focus:outline-none focus:border-accent focus:ring-[3px] focus:ring-accent/15"
            >
              {jwt.orgs.length === 0 && <option value="">No orgs</option>}
              {jwt.orgs.map((o) => (
                <option key={o} value={o}>Org #{o}</option>
              ))}
            </select>
            <input
              ref={searchRef}
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search by name or path…"
              disabled={importing}
              className="flex-1 bg-[#0e0e12] border border-border rounded-md px-3 py-2 text-[13px] text-ink placeholder:text-ink-mute focus:outline-none focus:border-accent focus:ring-[3px] focus:ring-accent/15"
            />
          </div>
        </div>

        {/* Body */}
        <div className="flex-1 overflow-y-auto px-3 py-3 min-h-[240px]">
          {state.kind === "loading" && (
            <div className="flex flex-col gap-1.5">
              {Array.from({ length: 5 }, (_, i) => (
                <div key={i} className="h-[52px] px-3 flex items-center gap-3">
                  <div className="shimmer-bar w-7 h-7 rounded-md" />
                  <div className="flex-1">
                    <div className="shimmer-bar h-3 w-[40%] mb-2" />
                    <div className="shimmer-bar h-[10px] w-[60%]" />
                  </div>
                </div>
              ))}
            </div>
          )}

          {state.kind === "success" && filtered.length > 0 && (
            <ul className="flex flex-col gap-1" role="listbox" aria-label="Importable projects">
              {filtered.map((c) => {
                const isImported = importedUrls.has(c.web_url);
                const isSelected = selected?.id === c.id;
                const isHttps = c.web_url.startsWith("https://");
                return (
                  <li key={c.id}>
                    <button
                      type="button"
                      role="option"
                      aria-selected={isSelected}
                      disabled={isImported || importing}
                      onClick={() => setSelected(c)}
                      className={[
                        "w-full grid grid-cols-[auto_1fr_auto] items-center gap-3 px-3 py-2.5 rounded-md text-left border transition-colors duration-100",
                        isImported
                          ? "border-transparent opacity-55 cursor-not-allowed"
                          : isSelected
                          ? "bg-accent/10 border-accent/40"
                          : "border-transparent hover:bg-[#1b1b22] hover:border-border",
                      ].join(" ")}
                    >
                      <span className="w-7 h-7 rounded-md bg-[#0e0e12] border border-border grid place-items-center text-ink-dim">
                        <HostIcon className="w-3.5 h-3.5" />
                      </span>
                      <span className="min-w-0">
                        <span className="block text-[13px] font-medium text-ink truncate leading-tight">{c.name}</span>
                        <span className="mt-0.5 flex items-center gap-1.5 font-mono text-[11px] text-ink-mute truncate leading-tight">
                          <LockIcon open={!isHttps} className={`w-2.5 h-2.5 flex-none ${isHttps ? "" : "text-[#b98a5a]"}`} />
                          <span className="truncate">{c.full_path}</span>
                        </span>
                      </span>
                      {isImported ? (
                        <span className="font-mono text-[10px] tracking-[0.08em] uppercase text-ink-mute px-1.5 py-0.5 border border-border rounded bg-[#0e0e12] whitespace-nowrap">
                          Imported
                        </span>
                      ) : isSelected ? (
                        <span className="inline-flex items-center justify-center w-4 h-4 rounded-full bg-accent text-accent-ink">
                          <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={3} strokeLinecap="round" strokeLinejoin="round"><path d="M5 12l5 5L20 7" /></svg>
                        </span>
                      ) : (
                        <span className="font-mono text-[10px] tracking-[0.08em] uppercase text-ink-mute whitespace-nowrap">org #{c.org}</span>
                      )}
                    </button>
                  </li>
                );
              })}
            </ul>
          )}

          {state.kind === "success" && filtered.length === 0 && (
            <div className="h-full grid place-items-center text-center px-6 py-10 text-[13px] text-ink-dim">
              No projects match "<span className="text-ink">{query}</span>".
            </div>
          )}

          {state.kind === "empty" && (
            <div className="h-full grid place-items-center text-center px-6 py-10">
              <div>
                <div className="mx-auto mb-3 w-9 h-9 rounded-lg border border-border bg-surface-2 grid place-items-center text-ink-mute">
                  <EmptyIcon className="w-[18px] h-[18px]" />
                </div>
                <p className="text-sm font-medium text-ink mb-1">No projects on this host</p>
                <p className="text-[13px] text-ink-dim">Nothing to import yet. Create one on {hostLabel} first.</p>
              </div>
            </div>
          )}

          {state.kind === "error" && (
            <div className="h-full grid place-items-center text-center px-6 py-10">
              <div>
                <div className="mx-auto mb-3 w-9 h-9 rounded-lg border border-[#3a2326] bg-surface-2 grid place-items-center text-danger">
                  <AlertIcon className="w-[18px] h-[18px]" />
                </div>
                <p className="text-sm font-medium text-ink mb-1">Couldn't fetch projects</p>
                <p className="text-[13px] text-ink-dim mb-3.5">{state.message || "We couldn't reach the host."}</p>
                <button
                  type="button"
                  onClick={() => fetchCandidates(org)}
                  className="inline-flex items-center gap-2 bg-transparent text-ink border border-border-strong px-3 py-1.5 rounded-md text-[13px] font-medium hover:bg-[#1b1b22] hover:border-[#3a3a44] active:translate-y-px"
                >
                  <RefreshIcon />Retry
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="px-6 py-4 border-t border-border flex items-center justify-between gap-3">
          <div className="min-w-0 text-[12px] text-ink-dim truncate">
            {importError ? (
              <span className="text-danger">{importError}</span>
            ) : selected ? (
              <span className="flex items-center gap-1.5">
                <ArrowIcon className="text-ink-mute" />
                <span className="text-ink-mute">Will import</span>
                <span className="text-ink font-medium truncate">{selected.full_path}</span>
              </span>
            ) : (
              <span>Select a project to import.</span>
            )}
          </div>
          <div className="flex gap-2 flex-none">
            <button
              type="button"
              onClick={onClose}
              disabled={importing}
              className="bg-transparent text-ink-dim border border-border-strong px-3 py-1.5 rounded-md text-[13px] font-medium whitespace-nowrap hover:bg-[#1b1b22] hover:text-ink active:translate-y-px disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={doImport}
              disabled={!selected || importing}
              className="inline-flex items-center gap-2 whitespace-nowrap bg-accent text-accent-ink border border-accent px-3.5 py-1.5 rounded-md text-[13px] font-semibold hover:brightness-105 active:translate-y-px disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {importing && <Spinner className="!border-accent-ink/30 !border-t-accent-ink" />}
              {importing ? "Importing…" : "Import project"}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
