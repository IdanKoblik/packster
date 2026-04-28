import { useEffect, useRef, useState } from "react";
import {
  UnauthorizedError,
  listPermissions,
  revokePermission,
  searchUserCandidates,
  setPermission,
} from "../api";
import type { PermissionEntry, UserCandidate } from "../types";
import { Spinner } from "./icons";

interface Props {
  token: string;
  projectId: number;
  projectName: string;
  onClose: () => void;
  onUnauthorized: () => void;
  onError: (message: string) => void;
  onSuccess: (message: string) => void;
}

type State =
  | { kind: "loading" }
  | { kind: "ready"; entries: PermissionEntry[] }
  | { kind: "error"; message: string };

export function PermissionsModal({
  token,
  projectId,
  projectName,
  onClose,
  onUnauthorized,
  onError,
  onSuccess,
}: Props) {
  const [state, setState] = useState<State>({ kind: "loading" });
  const [query, setQuery] = useState("");
  const [candidates, setCandidates] = useState<UserCandidate[]>([]);
  const [searching, setSearching] = useState(false);
  const [busyUserId, setBusyUserId] = useState<number | null>(null);
  const searchTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  const refresh = async () => {
    setState({ kind: "loading" });
    try {
      const entries = await listPermissions(token, projectId);
      setState({ kind: "ready", entries });
    } catch (err) {
      if (err instanceof UnauthorizedError) return onUnauthorized();
      setState({ kind: "error", message: err instanceof Error ? err.message : String(err) });
    }
  };

  useEffect(() => {
    refresh();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [projectId]);

  useEffect(() => {
    if (searchTimer.current) clearTimeout(searchTimer.current);
    const q = query.trim();
    if (!q) {
      setCandidates([]);
      setSearching(false);
      return;
    }
    setSearching(true);
    searchTimer.current = setTimeout(async () => {
      try {
        const results = await searchUserCandidates(token, projectId, q);
        const existing = state.kind === "ready" ? new Set(state.entries.map((e) => e.user_id)) : new Set<number>();
        setCandidates(results.filter((u) => !existing.has(u.id)));
      } catch (err) {
        if (err instanceof UnauthorizedError) return onUnauthorized();
        onError(err instanceof Error ? err.message : String(err));
      } finally {
        setSearching(false);
      }
    }, 250);
    return () => {
      if (searchTimer.current) clearTimeout(searchTimer.current);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query, projectId, token, state.kind]);

  const updateFlag = async (entry: PermissionEntry, key: "can_download" | "can_upload" | "can_delete", value: boolean) => {
    if (entry.is_owner) return;
    setBusyUserId(entry.user_id);
    try {
      await setPermission(token, projectId, {
        user_id: entry.user_id,
        can_download: key === "can_download" ? value : entry.can_download,
        can_upload: key === "can_upload" ? value : entry.can_upload,
        can_delete: key === "can_delete" ? value : entry.can_delete,
      });
      setState((s) =>
        s.kind === "ready"
          ? {
              kind: "ready",
              entries: s.entries.map((e) =>
                e.user_id === entry.user_id ? { ...e, [key]: value } : e,
              ),
            }
          : s,
      );
    } catch (err) {
      if (err instanceof UnauthorizedError) return onUnauthorized();
      onError(err instanceof Error ? err.message : String(err));
    } finally {
      setBusyUserId(null);
    }
  };

  const revoke = async (entry: PermissionEntry) => {
    if (entry.is_owner) return;
    if (!confirm(`Revoke ${entry.display_name}'s access to "${projectName}"?`)) return;
    setBusyUserId(entry.user_id);
    try {
      await revokePermission(token, projectId, entry.user_id);
      onSuccess(`Revoked access for ${entry.display_name}`);
      setState((s) =>
        s.kind === "ready"
          ? { kind: "ready", entries: s.entries.filter((e) => e.user_id !== entry.user_id) }
          : s,
      );
    } catch (err) {
      if (err instanceof UnauthorizedError) return onUnauthorized();
      onError(err instanceof Error ? err.message : String(err));
    } finally {
      setBusyUserId(null);
    }
  };

  const grant = async (candidate: UserCandidate) => {
    setBusyUserId(candidate.id);
    try {
      await setPermission(token, projectId, {
        user_id: candidate.id,
        can_download: true,
        can_upload: false,
        can_delete: false,
      });
      onSuccess(`Granted ${candidate.display_name} download access`);
      setQuery("");
      setCandidates([]);
      await refresh();
    } catch (err) {
      if (err instanceof UnauthorizedError) return onUnauthorized();
      onError(err instanceof Error ? err.message : String(err));
    } finally {
      setBusyUserId(null);
    }
  };

  return (
    <div
      className="fixed inset-0 z-40 flex items-center justify-center p-6 bg-black/60 backdrop-blur-sm animate-fade-in"
      onClick={onClose}
      role="dialog"
      aria-modal="true"
      aria-label="Manage permissions"
    >
      <div
        onClick={(e) => e.stopPropagation()}
        className="w-full max-w-[640px] bg-surface border border-border rounded-xl shadow-card flex flex-col max-h-[85vh]"
      >
        <div className="px-6 pt-6 pb-4 border-b border-border">
          <div className="inline-flex items-center gap-2 font-mono text-[11px] tracking-[0.14em] uppercase text-ink-mute mb-3.5">
            <span className="w-1.5 h-1.5 rounded-full bg-accent shadow-[0_0_0_3px_rgba(215,226,122,0.12)]" />
            Permissions
          </div>
          <h2 className="text-lg font-semibold tracking-[-0.01em] text-ink m-0 mb-1">Manage access</h2>
          <p className="text-[13px] text-ink-dim leading-relaxed m-0">
            Grant other users access to <span className="font-mono text-ink">{projectName}</span>.
          </p>
        </div>

        <div className="px-6 py-4 border-b border-border">
          <label className="block text-[12px] text-ink-dim mb-1.5">Add user</label>
          <div className="relative">
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="search by display name…"
              className="w-full bg-[#0e0e12] border border-border rounded-md px-3 py-2 text-[13px] text-ink placeholder:text-ink-mute focus:outline-none focus:border-accent focus:ring-[3px] focus:ring-accent/15"
            />
            {(candidates.length > 0 || (searching && query.trim())) && (
              <div className="absolute left-0 right-0 top-full mt-1 bg-surface-2 border border-border rounded-md shadow-card max-h-[220px] overflow-y-auto z-10">
                {searching && (
                  <div className="px-3 py-2 text-[12px] text-ink-mute inline-flex items-center gap-2">
                    <Spinner /> Searching…
                  </div>
                )}
                {!searching && candidates.length === 0 && (
                  <div className="px-3 py-2 text-[12px] text-ink-mute">No matches.</div>
                )}
                {candidates.map((c) => (
                  <button
                    key={c.id}
                    type="button"
                    onClick={() => grant(c)}
                    disabled={busyUserId === c.id}
                    className="w-full text-left px-3 py-2 text-[13px] text-ink hover:bg-[#1b1b22] disabled:opacity-50 inline-flex items-center justify-between"
                  >
                    <span>{c.display_name}</span>
                    <span className="font-mono text-[11px] text-ink-mute">#{c.id}</span>
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>

        <div className="flex-1 overflow-y-auto px-6 py-4">
          {state.kind === "loading" && (
            <div className="inline-flex items-center gap-2 text-[12px] text-ink-mute">
              <Spinner /> Loading permissions…
            </div>
          )}

          {state.kind === "error" && (
            <div className="text-[13px] text-danger">{state.message}</div>
          )}

          {state.kind === "ready" && (
            <div className="flex flex-col gap-2">
              {state.entries.map((e) => (
                <div
                  key={e.user_id}
                  className="bg-surface-2 border border-border rounded-lg px-3 py-2.5 grid grid-cols-[1fr_auto] items-center gap-3"
                >
                  <div className="min-w-0">
                    <div className="text-[13px] text-ink truncate">
                      {e.display_name}
                      {e.is_owner && (
                        <span className="ml-2 font-mono text-[10px] tracking-[0.1em] uppercase text-accent border border-accent/30 px-1.5 py-px rounded">
                          owner
                        </span>
                      )}
                    </div>
                    <div className="font-mono text-[11px] text-ink-mute">#{e.user_id}</div>
                  </div>
                  <div className="flex items-center gap-3">
                    <PermFlag
                      label="download"
                      checked={e.can_download}
                      disabled={e.is_owner || busyUserId === e.user_id}
                      onChange={(v) => updateFlag(e, "can_download", v)}
                    />
                    <PermFlag
                      label="upload"
                      checked={e.can_upload}
                      disabled={e.is_owner || busyUserId === e.user_id}
                      onChange={(v) => updateFlag(e, "can_upload", v)}
                    />
                    <PermFlag
                      label="delete"
                      checked={e.can_delete}
                      disabled={e.is_owner || busyUserId === e.user_id}
                      onChange={(v) => updateFlag(e, "can_delete", v)}
                    />
                    {!e.is_owner && (
                      <button
                        type="button"
                        onClick={() => revoke(e)}
                        disabled={busyUserId === e.user_id}
                        className="bg-transparent text-danger border border-[#3a2326] px-2.5 py-1 rounded-md text-[12px] font-medium hover:bg-[#2a1619] active:translate-y-px disabled:opacity-50"
                      >
                        Revoke
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="px-6 py-4 border-t border-border flex items-center justify-end gap-2">
          <button
            type="button"
            onClick={onClose}
            className="bg-transparent text-ink-dim border border-border-strong px-3 py-1.5 rounded-md text-[13px] font-medium whitespace-nowrap hover:bg-[#1b1b22] hover:text-ink active:translate-y-px"
          >
            Done
          </button>
        </div>
      </div>
    </div>
  );
}

interface PermFlagProps {
  label: string;
  checked: boolean;
  disabled: boolean;
  onChange: (value: boolean) => void;
}

function PermFlag({ label, checked, disabled, onChange }: PermFlagProps) {
  return (
    <label
      className={`inline-flex items-center gap-1.5 text-[12px] ${disabled ? "opacity-50" : "cursor-pointer hover:text-ink"} text-ink-dim`}
    >
      <input
        type="checkbox"
        checked={checked}
        disabled={disabled}
        onChange={(e) => onChange(e.target.checked)}
        className="accent-accent"
      />
      {label}
    </label>
  );
}
