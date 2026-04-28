import { useCallback, useEffect, useRef, useState } from "react";
import type { Jwt, Project, ProjectsState, Toast } from "./types";
import { clearJwt } from "./auth";
import { parseHost } from "./lib";
import { getProvider } from "./providers";
import { ProjectRow } from "./components/ProjectRow";
import { StateBlock } from "./components/StateBlock";
import { CreateProjectModal } from "./components/CreateProjectModal";
import { ToastStack } from "./components/ToastStack";
import { AlertIcon, EmptyIcon, PlusIcon } from "./components/icons";

interface Props {
  jwt: Jwt;
  raw: string;
  onLoggedOut: () => void;
  onOpenProject: (project: Project) => void;
}

function SkeletonRow() {
  return (
    <div className="bg-surface-2 border border-border rounded-lg px-3.5 py-3 grid grid-cols-[auto_1fr_auto] items-center gap-3" aria-hidden="true">
      <div className="shimmer-bar w-8 h-8 rounded-md" />
      <div>
        <div className="shimmer-bar h-3 w-[50%] mb-2" />
        <div className="shimmer-bar h-[10px] w-[70%]" />
      </div>
      <div className="shimmer-bar h-4 w-14 rounded" />
    </div>
  );
}

export function Dashboard({ jwt, raw, onLoggedOut, onOpenProject }: Props) {
  const [state, setState] = useState<ProjectsState>({ kind: "loading" });
  const [showCreate, setShowCreate] = useState(false);
  const [toasts, setToasts] = useState<Toast[]>([]);
  const toastId = useRef(0);

  const pushToast = useCallback((kind: Toast["kind"], message: string) => {
    const id = ++toastId.current;
    setToasts((ts) => [...ts, { id, kind, message }]);
    setTimeout(() => setToasts((ts) => ts.filter((t) => t.id !== id)), 4500);
  }, []);

  const dismissToast = (id: number) =>
    setToasts((ts) => ts.filter((t) => t.id !== id));

  const handleUnauthorized = useCallback(() => {
    pushToast("error", "Your session expired. Redirecting to sign in…");
    setTimeout(() => {
      clearJwt();
      onLoggedOut();
    }, 1200);
  }, [onLoggedOut, pushToast]);

  const fetchProjects = useCallback(async () => {
    setState({ kind: "loading" });
    try {
      const provider = getProvider(jwt.host.type);
      const projects = await provider.fetchUserProjects({ token: raw, orgs: jwt.orgs });
      if (!projects.length) setState({ kind: "empty" });
      else setState({ kind: "success", projects });
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      if (/401|403/.test(msg)) return handleUnauthorized();
      setState({ kind: "error", message: msg });
    }
  }, [jwt, handleUnauthorized]);

  useEffect(() => {
    fetchProjects();
  }, [fetchProjects]);

  const handleLogout = () => {
    clearJwt();
    onLoggedOut();
  };

  const handleImported = (project?: Project) => {
    setShowCreate(false);
    pushToast("success", "Project imported");
    if (project && state.kind === "success") {
      setState({ kind: "success", projects: [project, ...state.projects] });
    } else {
      fetchProjects();
    }
  };

  const importedUrls = new Set(
    state.kind === "success" ? state.projects.map((p) => p.web_url) : []
  );

  const hostLabel = parseHost(jwt.host.url).host || jwt.host.url;

  return (
    <div className="relative min-h-screen flex flex-col">
      {/* Top bar */}
      <header className="relative z-10 flex items-center justify-between px-7 py-4 border-b border-border bg-bg/60 backdrop-blur">
        <div className="flex items-center gap-4">
          <span className="text-ink font-semibold text-sm tracking-[-0.01em]">packster</span>
          <span className="hidden sm:inline-block w-px h-4 bg-border" />
          <span className="hidden sm:inline-flex items-center gap-1.5 font-mono text-[11px] tracking-[0.08em] uppercase text-ink-mute">
            <span className="w-1.5 h-1.5 rounded-full bg-ok shadow-[0_0_0_3px_rgba(107,208,122,0.14)]" />
            {hostLabel}
          </span>
        </div>

        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => setShowCreate(true)}
            className="inline-flex items-center gap-1.5 whitespace-nowrap bg-accent text-accent-ink border border-accent px-3 py-1.5 rounded-md text-[13px] font-semibold hover:brightness-105 active:translate-y-px"
            aria-label="Create project"
          >
            <PlusIcon />
            Import project
          </button>
          <button
            type="button"
            onClick={handleLogout}
            className="bg-transparent text-ink-dim border border-border-strong px-3 py-1.5 rounded-md text-[13px] font-medium whitespace-nowrap hover:bg-[#1b1b22] hover:text-ink active:translate-y-px"
          >
            Log out
          </button>
        </div>
      </header>

      <main className="relative z-10 flex-1 px-7 py-8 max-w-[1120px] w-full mx-auto">
        <div className="flex items-baseline justify-between mb-5">
          <div>
            <h1 className="text-xl font-semibold tracking-[-0.01em] text-ink m-0 mb-1">Your projects</h1>
            <p className="text-[13px] text-ink-dim m-0">
              Imported across {jwt.orgs.length} org{jwt.orgs.length === 1 ? "" : "s"}.
            </p>
          </div>
          {state.kind === "success" && (
            <span className="font-mono text-[11px] tracking-[0.08em] uppercase text-ink-mute tabular-nums">
              {state.projects.length} {state.projects.length === 1 ? "project" : "projects"}
            </span>
          )}
        </div>

        <div className="animate-fade-in" key={state.kind}>
          {state.kind === "loading" && (
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              {Array.from({ length: 6 }, (_, i) => <SkeletonRow key={i} />)}
            </div>
          )}

          {state.kind === "success" && (
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              {state.projects.map((p) => (
                <ProjectRow
                  key={p.id}
                  project={p}
                  hostType={jwt.host.type}
                  onOpen={() => onOpenProject(p)}
                />
              ))}
            </div>
          )}

          {state.kind === "empty" && (
            <StateBlock
              icon={<EmptyIcon />}
              title="No projects yet"
              actionLabel="Import your first project"
              onAction={() => setShowCreate(true)}
            >
              You haven't imported any projects yet. Pull one in from {hostLabel}.
            </StateBlock>
          )}

          {state.kind === "error" && (
            <StateBlock
              icon={<AlertIcon />}
              iconTone="danger"
              title="Couldn't load projects"
              actionLabel="Retry"
              onAction={fetchProjects}
            >
              {state.message || "We couldn't reach the server. Check your connection and try again."}
            </StateBlock>
          )}
        </div>
      </main>

      {showCreate && (
        <CreateProjectModal
          jwt={jwt}
          raw={raw}
          importedUrls={importedUrls}
          onClose={() => setShowCreate(false)}
          onImported={(p) => handleImported(p as Project)}
          onUnauthorized={() => {
            setShowCreate(false);
            handleUnauthorized();
          }}
        />
      )}

      <ToastStack toasts={toasts} onDismiss={dismissToast} />
    </div>
  );
}
