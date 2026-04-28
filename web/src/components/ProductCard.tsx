import { useEffect, useState } from "react";
import type { Permission, Product, Version, VersionsState } from "../types";
import { listVersions, UnauthorizedError } from "../api";
import { VersionRow } from "./VersionRow";
import { AlertIcon, EmptyIcon, PlusIcon, RefreshIcon, Spinner } from "./icons";

interface Props {
  product: Product;
  permission: Permission;
  token: string;
  defaultExpanded?: boolean;
  onUploadClick: (product: Product) => void;
  onDeleteProduct: (product: Product) => Promise<void>;
  onDownloadVersion: (version: Version) => Promise<void>;
  onDeleteVersion: (version: Version) => Promise<void>;
  onUnauthorized: () => void;
  reloadKey: number;
}

function Chevron({ open }: { open: boolean }) {
  return (
    <svg
      viewBox="0 0 24 24"
      width="14"
      height="14"
      fill="none"
      stroke="currentColor"
      strokeWidth={2}
      strokeLinecap="round"
      strokeLinejoin="round"
      className={`transition-transform duration-150 ${open ? "rotate-90" : ""}`}
      aria-hidden="true"
    >
      <path d="M9 6l6 6-6 6" />
    </svg>
  );
}

export function ProductCard({
  product,
  permission,
  token,
  defaultExpanded = false,
  onUploadClick,
  onDeleteProduct,
  onDownloadVersion,
  onDeleteVersion,
  onUnauthorized,
  reloadKey,
}: Props) {
  const [expanded, setExpanded] = useState(defaultExpanded);
  const [state, setState] = useState<VersionsState>({ kind: "loading" });
  const [deleting, setDeleting] = useState(false);
  const [hasLoaded, setHasLoaded] = useState(false);

  useEffect(() => {
    if (!expanded) return;
    let cancelled = false;
    setState({ kind: "loading" });
    (async () => {
      try {
        const versions = await listVersions(token, product.id);
        if (cancelled) return;
        setHasLoaded(true);
        if (!versions.length) setState({ kind: "empty" });
        else setState({ kind: "success", versions });
      } catch (err) {
        if (cancelled) return;
        if (err instanceof UnauthorizedError) return onUnauthorized();
        setState({ kind: "error", message: err instanceof Error ? err.message : String(err) });
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [expanded, token, product.id, reloadKey, onUnauthorized]);

  const doDeleteProduct = async (e: React.MouseEvent) => {
    e.stopPropagation();
    if (deleting) return;
    if (!confirm(`Delete product "${product.name}" and all its versions?`)) return;
    setDeleting(true);
    try {
      await onDeleteProduct(product);
    } finally {
      setDeleting(false);
    }
  };

  const handleUpload = (e: React.MouseEvent) => {
    e.stopPropagation();
    onUploadClick(product);
  };

  const versionsCountBadge = hasLoaded && state.kind === "success"
    ? `${state.versions.length} ${state.versions.length === 1 ? "version" : "versions"}`
    : hasLoaded && state.kind === "empty"
    ? "no versions"
    : null;

  return (
    <div className="bg-surface-2 border border-border rounded-lg overflow-hidden">
      <button
        type="button"
        onClick={() => setExpanded((x) => !x)}
        aria-expanded={expanded}
        className="w-full px-4 py-3 flex items-center gap-3 text-left hover:bg-[#1b1b22] focus:outline-none focus-visible:bg-[#1b1b22]"
      >
        <span className="text-ink-mute"><Chevron open={expanded} /></span>
        <div className="flex-1 min-w-0">
          <div className="text-sm font-medium text-ink truncate">{product.name}</div>
          <div className="font-mono text-[11px] text-ink-mute">
            product #{product.id}
            {versionsCountBadge && <> • {versionsCountBadge}</>}
          </div>
        </div>
        {permission.can_upload && (
          <span
            role="button"
            tabIndex={0}
            onClick={handleUpload}
            onKeyDown={(e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                onUploadClick(product);
              }
            }}
            className="inline-flex items-center gap-1.5 whitespace-nowrap bg-accent text-accent-ink border border-accent px-2.5 py-1 rounded-md text-[12px] font-semibold hover:brightness-105 active:translate-y-px"
          >
            <PlusIcon />
            New version
          </span>
        )}
        {permission.can_delete && (
          <span
            role="button"
            tabIndex={0}
            onClick={doDeleteProduct}
            onKeyDown={(e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                doDeleteProduct(e as unknown as React.MouseEvent);
              }
            }}
            aria-disabled={deleting}
            className={`inline-flex items-center gap-1.5 bg-transparent text-danger border border-[#3a2326] px-2.5 py-1 rounded-md text-[12px] font-medium hover:bg-[#2a1619] active:translate-y-px ${deleting ? "opacity-50 pointer-events-none" : ""}`}
          >
            {deleting ? "Deleting…" : "Delete"}
          </span>
        )}
      </button>

      {expanded && (
        <div className="border-t border-border p-2">
          {state.kind === "loading" && (
            <div className="px-3 py-3 inline-flex items-center gap-2 text-[12px] text-ink-mute">
              <Spinner /> Loading versions…
            </div>
          )}

          {state.kind === "success" && (
            <div className="flex flex-col">
              {state.versions.map((v) => (
                <VersionRow
                  key={v.id}
                  version={v}
                  permission={permission}
                  onDownload={() => onDownloadVersion(v)}
                  onDelete={() => onDeleteVersion(v)}
                />
              ))}
            </div>
          )}

          {state.kind === "empty" && (
            <div className="px-3 py-5 grid place-items-center text-center">
              <div className="mx-auto mb-2 w-7 h-7 rounded-md border border-border bg-surface grid place-items-center text-ink-mute">
                <EmptyIcon className="w-[14px] h-[14px]" />
              </div>
              <p className="text-[13px] text-ink-dim m-0 mb-2">No versions yet. Each upload is one version.</p>
              {permission.can_upload && (
                <button
                  type="button"
                  onClick={() => onUploadClick(product)}
                  className="inline-flex items-center gap-1.5 bg-transparent text-ink border border-border-strong px-2.5 py-1 rounded-md text-[12px] font-medium hover:bg-[#1b1b22]"
                >
                  <PlusIcon />
                  Upload first version
                </button>
              )}
            </div>
          )}

          {state.kind === "error" && (
            <div className="px-3 py-5 grid place-items-center text-center">
              <div className="mx-auto mb-2 w-7 h-7 rounded-md border border-[#3a2326] bg-surface grid place-items-center text-danger">
                <AlertIcon className="w-[14px] h-[14px]" />
              </div>
              <p className="text-[13px] text-ink-dim mb-2">{state.message || "Couldn't load versions."}</p>
              <button
                type="button"
                onClick={() => setState({ kind: "loading" })}
                className="inline-flex items-center gap-1.5 bg-transparent text-ink border border-border-strong px-2.5 py-1 rounded-md text-[12px] font-medium hover:bg-[#1b1b22]"
              >
                <RefreshIcon />
                Retry
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
