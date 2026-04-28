import { useCallback, useEffect, useRef, useState } from "react";
import {
  UnauthorizedError,
  createProduct,
  deleteProduct,
  deleteProject,
  deleteVersion,
  downloadVersion,
  listProducts,
  uploadVersion,
} from "./api";
import { CreateProductModal } from "./components/CreateProductModal";
import { UploadVersionModal } from "./components/UploadVersionModal";
import { PermissionsModal } from "./components/PermissionsModal";
import { ProductCard } from "./components/ProductCard";
import { StateBlock } from "./components/StateBlock";
import { ToastStack } from "./components/ToastStack";
import { AlertIcon, ArrowIcon, EmptyIcon, PlusIcon } from "./components/icons";
import type { Jwt, Permission, Product, ProductsState, Project, Toast, Version } from "./types";
import { parseHost } from "./lib";

interface Props {
  jwt: Jwt;
  raw: string;
  project: Project;
  onBack: () => void;
  onUnauthorized: () => void;
  onProjectDeleted?: (projectId: number) => void;
}

const FULL_ACCESS: Permission = {
  account: 0,
  project: 0,
  can_download: true,
  can_upload: true,
  can_delete: true,
};

export function ProjectDetail({ jwt, raw, project, onBack, onUnauthorized, onProjectDeleted }: Props) {
  const [state, setState] = useState<ProductsState>({ kind: "loading" });
  const [showCreate, setShowCreate] = useState(false);
  const [uploadFor, setUploadFor] = useState<Product | null>(null);
  const [reloadKey, setReloadKey] = useState(0);
  const [toasts, setToasts] = useState<Toast[]>([]);
  const [showPermissions, setShowPermissions] = useState(false);
  const [deletingProject, setDeletingProject] = useState(false);
  const toastId = useRef(0);

  const currentUserId = jwt.sub ? Number(jwt.sub) : NaN;
  const isOwner =
    typeof project.owner === "number" &&
    !Number.isNaN(currentUserId) &&
    project.owner === currentUserId;

  const pushToast = useCallback((kind: Toast["kind"], message: string) => {
    const id = ++toastId.current;
    setToasts((ts) => [...ts, { id, kind, message }]);
    setTimeout(() => setToasts((ts) => ts.filter((t) => t.id !== id)), 4500);
  }, []);

  const permission: Permission = FULL_ACCESS;

  const fetchProducts = useCallback(async () => {
    setState({ kind: "loading" });
    try {
      const products = await listProducts(raw, project.id);
      if (!products.length) setState({ kind: "empty" });
      else setState({ kind: "success", products });
    } catch (err) {
      if (err instanceof UnauthorizedError) return onUnauthorized();
      setState({ kind: "error", message: err instanceof Error ? err.message : String(err) });
    }
  }, [raw, project.id, onUnauthorized]);

  useEffect(() => {
    fetchProducts();
  }, [fetchProducts]);

  const handleCreate = async (name: string) => {
    try {
      const product = await createProduct(raw, project.id, name);
      setShowCreate(false);
      pushToast("success", `Product "${product.name}" created`);
      setState((s) => {
        if (s.kind === "success") return { kind: "success", products: [...s.products, product] };
        return { kind: "success", products: [product] };
      });
    } catch (err) {
      if (err instanceof UnauthorizedError) return onUnauthorized();
      throw err;
    }
  };

  const handleDeleteProduct = async (product: Product) => {
    try {
      await deleteProduct(raw, project.id, product.id);
      pushToast("success", `Product "${product.name}" deleted`);
      setState((s) => {
        if (s.kind !== "success") return s;
        const next = s.products.filter((p) => p.id !== product.id);
        return next.length ? { kind: "success", products: next } : { kind: "empty" };
      });
    } catch (err) {
      if (err instanceof UnauthorizedError) return onUnauthorized();
      pushToast("error", err instanceof Error ? err.message : String(err));
    }
  };

  const handleUpload = async (name: string, file: File) => {
    if (!uploadFor) return;
    try {
      const version = await uploadVersion(raw, uploadFor.id, name, file);
      setUploadFor(null);
      pushToast("success", `Uploaded ${version.name}`);
      setReloadKey((k) => k + 1);
    } catch (err) {
      if (err instanceof UnauthorizedError) return onUnauthorized();
      throw err;
    }
  };

  const handleDownloadVersion = async (version: Version) => {
    try {
      await downloadVersion(raw, version.id, version.path);
    } catch (err) {
      if (err instanceof UnauthorizedError) return onUnauthorized();
      pushToast("error", err instanceof Error ? err.message : String(err));
    }
  };

  const handleDeleteProject = async () => {
    if (deletingProject) return;
    if (!confirm(`Delete project "${project.name}"? All products, versions, and uploaded files will be removed.`)) return;
    setDeletingProject(true);
    try {
      await deleteProject(raw, project.id);
      onProjectDeleted?.(project.id);
      onBack();
    } catch (err) {
      if (err instanceof UnauthorizedError) return onUnauthorized();
      pushToast("error", err instanceof Error ? err.message : String(err));
      setDeletingProject(false);
    }
  };

  const handleDeleteVersion = async (version: Version) => {
    try {
      await deleteVersion(raw, version.id);
      pushToast("success", `Version ${version.path} deleted`);
      setReloadKey((k) => k + 1);
    } catch (err) {
      if (err instanceof UnauthorizedError) return onUnauthorized();
      pushToast("error", err instanceof Error ? err.message : String(err));
    }
  };

  const dismissToast = (id: number) => setToasts((ts) => ts.filter((t) => t.id !== id));
  const hostLabel = parseHost(jwt.host.url).host || jwt.host.url;

  return (
    <div className="relative min-h-screen flex flex-col">
      <header className="relative z-10 flex items-center justify-between px-7 py-4 border-b border-border bg-bg/60 backdrop-blur">
        <div className="flex items-center gap-4 min-w-0">
          <button
            type="button"
            onClick={onBack}
            className="inline-flex items-center gap-1.5 bg-transparent text-ink-dim border border-border-strong px-2.5 py-1 rounded-md text-[12px] font-medium whitespace-nowrap hover:bg-[#1b1b22] hover:text-ink active:translate-y-px"
            aria-label="Back to projects"
          >
            <ArrowIcon className="rotate-180" />
            Projects
          </button>
          <span className="hidden sm:inline-block w-px h-4 bg-border" />
          <div className="min-w-0">
            <div className="text-ink font-semibold text-sm tracking-[-0.01em] truncate">{project.name}</div>
            <div className="font-mono text-[11px] text-ink-mute truncate">
              {hostLabel} • org #{project.org}
            </div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <a
            href={project.web_url}
            target="_blank"
            rel="noreferrer"
            className="inline-flex items-center gap-1.5 bg-transparent text-ink-dim border border-border-strong px-3 py-1.5 rounded-md text-[13px] font-medium whitespace-nowrap hover:bg-[#1b1b22] hover:text-ink no-underline"
          >
            View on {jwt.host.type === "github" ? "GitHub" : "GitLab"}
          </a>
          {isOwner && (
            <button
              type="button"
              onClick={() => setShowPermissions(true)}
              className="bg-transparent text-ink-dim border border-border-strong px-3 py-1.5 rounded-md text-[13px] font-medium whitespace-nowrap hover:bg-[#1b1b22] hover:text-ink active:translate-y-px"
            >
              Permissions
            </button>
          )}
          {isOwner && (
            <button
              type="button"
              onClick={handleDeleteProject}
              disabled={deletingProject}
              className="bg-transparent text-danger border border-[#3a2326] px-3 py-1.5 rounded-md text-[13px] font-medium whitespace-nowrap hover:bg-[#2a1619] active:translate-y-px disabled:opacity-50"
            >
              {deletingProject ? "Deleting…" : "Delete project"}
            </button>
          )}
          {permission.can_upload && (
            <button
              type="button"
              onClick={() => setShowCreate(true)}
              className="inline-flex items-center gap-1.5 whitespace-nowrap bg-accent text-accent-ink border border-accent px-3 py-1.5 rounded-md text-[13px] font-semibold hover:brightness-105 active:translate-y-px"
            >
              <PlusIcon />
              New product
            </button>
          )}
        </div>
      </header>

      <main className="relative z-10 flex-1 px-7 py-8 max-w-[1120px] w-full mx-auto">
        <div className="flex items-baseline justify-between mb-5">
          <div>
            <h1 className="text-xl font-semibold tracking-[-0.01em] text-ink m-0 mb-1">Products</h1>
            <p className="text-[13px] text-ink-dim m-0">
              Group versioned uploads under a name (e.g. <span className="font-mono">spigot</span>).
            </p>
          </div>
          {state.kind === "success" && (
            <span className="font-mono text-[11px] tracking-[0.08em] uppercase text-ink-mute tabular-nums">
              {state.products.length} {state.products.length === 1 ? "product" : "products"}
            </span>
          )}
        </div>

        <div className="animate-fade-in" key={state.kind}>
          {state.kind === "loading" && (
            <div className="grid grid-cols-1 gap-3">
              {Array.from({ length: 2 }, (_, i) => (
                <div key={i} className="h-[120px] rounded-lg border border-border bg-surface-2">
                  <div className="shimmer-bar h-3 w-[35%] m-4" />
                  <div className="shimmer-bar h-2 w-[20%] mx-4" />
                </div>
              ))}
            </div>
          )}

          {state.kind === "success" && (
            <div className="grid grid-cols-1 gap-3">
              {state.products.map((p) => (
                <ProductCard
                  key={p.id}
                  product={p}
                  permission={permission}
                  token={raw}
                  defaultExpanded={state.products.length === 1}
                  onUploadClick={setUploadFor}
                  onDeleteProduct={handleDeleteProduct}
                  onDownloadVersion={handleDownloadVersion}
                  onDeleteVersion={handleDeleteVersion}
                  onUnauthorized={onUnauthorized}
                  reloadKey={reloadKey}
                />
              ))}
            </div>
          )}

          {state.kind === "empty" && (
            <StateBlock
              icon={<EmptyIcon />}
              title="No products yet"
              actionLabel={permission.can_upload ? "Create your first product" : undefined}
              onAction={permission.can_upload ? () => setShowCreate(true) : undefined}
            >
              Products group uploaded artifacts (e.g. server jars, plugins) under this project.
            </StateBlock>
          )}

          {state.kind === "error" && (
            <StateBlock
              icon={<AlertIcon />}
              iconTone="danger"
              title="Couldn't load products"
              actionLabel="Retry"
              onAction={fetchProducts}
            >
              {state.message || "We couldn't reach the server. Check your connection and try again."}
            </StateBlock>
          )}
        </div>
      </main>

      {showCreate && <CreateProductModal onCancel={() => setShowCreate(false)} onSubmit={handleCreate} />}
      {uploadFor && (
        <UploadVersionModal
          product={uploadFor}
          onCancel={() => setUploadFor(null)}
          onSubmit={handleUpload}
        />
      )}
      {showPermissions && (
        <PermissionsModal
          token={raw}
          projectId={project.id}
          projectName={project.name}
          onClose={() => setShowPermissions(false)}
          onUnauthorized={onUnauthorized}
          onError={(m) => pushToast("error", m)}
          onSuccess={(m) => pushToast("success", m)}
        />
      )}

      <ToastStack toasts={toasts} onDismiss={dismissToast} />
    </div>
  );
}
