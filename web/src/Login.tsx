import { useCallback, useEffect, useRef, useState } from "react";
import type { DebugState, HealthState, Host, LoadState } from "./types";
import { MOCK_HOSTS, isLocalhost } from "./lib";
import { HostRow } from "./components/HostRow";
import { SkeletonList } from "./components/SkeletonList";
import { StateBlock } from "./components/StateBlock";
import { HealthIndicator } from "./components/HealthIndicator";
import { DebugStateSwitch } from "./components/DebugStateSwitch";
import { AlertIcon, EmptyIcon } from "./components/icons";

export function Login() {
  const [load, setLoad] = useState<LoadState>({ kind: "loading" });
  const [health, setHealth] = useState<{ state: HealthState; label: string }>({
    state: "pending",
    label: "Checking status…",
  });
  const [selected, setSelected] = useState<Host | null>(null);
  const [debugVisible] = useState(isLocalhost());
  const [debugState, setDebugState] = useState<DebugState>("success");
  const hasBooted = useRef(false);

  const fetchHosts = useCallback(async () => {
    setLoad({ kind: "loading" });
    try {
      const res = await fetch("/api/hosts", { headers: { accept: "application/json" } });
      if (!res.ok) throw new Error("bad response");
      const data = await res.json();
      const hosts: Host[] = Array.isArray(data) ? data : data.hosts ?? [];
      if (!hosts.length) return setLoad({ kind: "empty" });
      setLoad({ kind: "success", hosts });
    } catch {
      setTimeout(() => setLoad({ kind: "success", hosts: MOCK_HOSTS }), 400);
    }
  }, []);

  const checkHealth = useCallback(async () => {
    try {
      const res = await fetch("/api/health", { headers: { accept: "application/json" } });
      if (res.status === 200) return setHealth({ state: "ok", label: "All systems normal" });
      setHealth({ state: "down", label: "Service degraded" });
    } catch {
      setHealth({ state: "down", label: "Service unreachable" });
    }
  }, []);

  useEffect(() => {
    if (hasBooted.current) return;
    hasBooted.current = true;
    fetchHosts();
    checkHealth();
  }, [fetchHosts, checkHealth]);

  const handleDebugChange = (s: DebugState) => {
    setDebugState(s);
    setSelected(null);
    if (s === "success") {
      setLoad({ kind: "loading" });
      setTimeout(() => setLoad({ kind: "success", hosts: MOCK_HOSTS }), 450);
    } else if (s === "loading") {
      setLoad({ kind: "loading" });
    } else if (s === "empty") {
      setLoad({ kind: "empty" });
    } else {
      setLoad({ kind: "error" });
    }
  };

  const onSelect = (host: Host) => {
    setSelected(host);
    window.location.href = `/api/auth/${host.type}/redirect?id=${host.id}`;
  };

  return (
    <div className="relative min-h-screen flex flex-col">
      <div className="absolute top-6 left-7 right-7 flex justify-between items-center text-[13px] text-ink-mute z-10">
        <div className="inline-flex items-center gap-2.5 text-ink font-semibold tracking-[-0.01em]">
          <span className="text-sm">packster</span>
        </div>
        <div className="flex gap-[18px]">
          <a href="#" className="text-ink-mute hover:text-ink-dim no-underline">Docs</a>
        </div>
      </div>

      <main className="relative z-10 flex-1 grid place-items-center px-6 py-12">
        <section
          aria-live="polite"
          className="w-full max-w-[440px] bg-surface border border-border rounded-xl px-9 pt-9 pb-7 shadow-card"
        >
          <div className="inline-flex items-center gap-2 font-mono text-[11px] tracking-[0.14em] uppercase text-ink-mute mb-3.5">
            <span className="w-1.5 h-1.5 rounded-full bg-accent shadow-[0_0_0_3px_rgba(215,226,122,0.12)]" />
            Sign in
          </div>

          <h1 className="text-2xl font-semibold tracking-[-0.02em] m-0 mb-1.5 text-ink">
            Welcome to Packster
          </h1>
          <p className="text-sm text-ink-dim mb-6 leading-relaxed">
            Choose a provider to continue. You'll be redirected to authorize access, then returned here.
          </p>

          <div className="flex justify-between items-center font-mono text-[11px] tracking-[0.12em] uppercase text-ink-mute mb-2.5">
            <span>Providers</span>
            {load.kind === "success" && (
              <span className="tabular-nums">{load.hosts.length} configured</span>
            )}
            {load.kind === "empty" && <span className="tabular-nums">0 configured</span>}
          </div>

          <div className="animate-fade-in" key={load.kind}>
            {load.kind === "loading" && <SkeletonList count={3} />}

            {load.kind === "success" && (
              <div className="flex flex-col gap-2">
                {load.hosts.map((host, i) => (
                  <HostRow
                    key={`${host.type}-${host.url}-${i}`}
                    host={host}
                    disabledByPeer={selected !== null && selected !== host}
                    onSelect={onSelect}
                  />
                ))}
              </div>
            )}

            {load.kind === "empty" && (
              <StateBlock
                icon={<EmptyIcon />}
                title="No providers configured"
                actionLabel="Check again"
                onAction={fetchHosts}
              >
                Your administrator hasn't added any Git hosts yet. Add one to{" "}
                <code className="font-mono text-xs bg-[#0e0e12] px-1.5 py-0.5 border border-border rounded text-ink-dim">
                  config.hosts
                </code>{" "}
                and restart the service to get started.
              </StateBlock>
            )}

            {load.kind === "error" && (
              <StateBlock
                icon={<AlertIcon />}
                iconTone="danger"
                title="Couldn't reach the server"
                actionLabel="Retry"
                onAction={fetchHosts}
              >
                We weren't able to load providers from{" "}
                <code className="font-mono text-xs bg-[#0e0e12] px-1.5 py-0.5 border border-border rounded text-ink-dim">
                  /api/hosts
                </code>
                . Check your connection, then try again.
              </StateBlock>
            )}
          </div>

          <div
            className="text-center text-xs text-ink-mute"
            style={{ marginTop: "22px", paddingTop: "18px", borderTop: "1px solid #24242b" }}
          >
            <span>
              New to Packster?&nbsp;{" "}
              <a href="#" className="text-ink-dim hover:text-ink no-underline">
                Read the setup guide
              </a>
            </span>
          </div>
        </section>
      </main>

      <footer className="relative z-10 px-7 py-6 flex justify-between items-center text-xs text-ink-mute">
        <HealthIndicator state={health.state} label={health.label} />
        <div className="flex gap-5">
          <a href="#" className="text-ink-mute hover:text-ink-dim no-underline">Documentation</a>
          <a href="#" className="text-ink-mute hover:text-ink-dim no-underline">Privacy</a>
        </div>
        <div>© Packster</div>
      </footer>

      {debugVisible && <DebugStateSwitch state={debugState} onChange={handleDebugChange} />}
    </div>
  );
}
