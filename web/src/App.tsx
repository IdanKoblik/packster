import { useEffect, useState } from "react";
import { Login } from "./Login";
import { Dashboard } from "./Dashboard";
import { ProjectDetail } from "./ProjectDetail";
import { clearJwt, getStoredJwt, verifySession } from "./auth";
import type { Jwt, Project } from "./types";

type Session = { raw: string; payload: Jwt } | null;
type View = { kind: "dashboard" } | { kind: "project"; project: Project };

export function App() {
  const [session, setSession] = useState<Session>(() => getStoredJwt());
  const [verifying, setVerifying] = useState<boolean>(() => getStoredJwt() !== null);
  const [view, setView] = useState<View>({ kind: "dashboard" });

  useEffect(() => {
    const onStorage = () => setSession(getStoredJwt());
    window.addEventListener("storage", onStorage);
    return () => window.removeEventListener("storage", onStorage);
  }, []);

  useEffect(() => {
    if (!session) {
      setVerifying(false);
      return;
    }
    let cancelled = false;
    setVerifying(true);
    verifySession(session.raw).then((ok) => {
      if (cancelled) return;
      if (!ok) {
        clearJwt();
        setSession(null);
      }
      setVerifying(false);
    });
    return () => {
      cancelled = true;
    };
  }, [session]);

  const onLoggedOut = () => {
    setSession(null);
    setView({ kind: "dashboard" });
  };

  if (verifying) return null;
  if (!session) return <Login />;
  if (view.kind === "project") {
    return (
      <ProjectDetail
        jwt={session.payload}
        raw={session.raw}
        project={view.project}
        onBack={() => setView({ kind: "dashboard" })}
        onUnauthorized={() => {
          clearJwt();
          onLoggedOut();
        }}
      />
    );
  }
  return (
    <Dashboard
      jwt={session.payload}
      raw={session.raw}
      onLoggedOut={onLoggedOut}
      onOpenProject={(project) => setView({ kind: "project", project })}
    />
  );
}
