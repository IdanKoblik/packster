import { useEffect, useRef, useState } from "react";
import type { Product } from "../types";
import { Spinner } from "./icons";

interface Props {
  product: Product;
  onCancel: () => void;
  onSubmit: (name: string, file: File) => Promise<void>;
}

function formatBytes(n: number): string {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  if (n < 1024 * 1024 * 1024) return `${(n / 1024 / 1024).toFixed(1)} MB`;
  return `${(n / 1024 / 1024 / 1024).toFixed(2)} GB`;
}

const VALID_NAME = /^[A-Za-z0-9._+-]+$/;

export function UploadVersionModal({ product, onCancel, onSubmit }: Props) {
  const [name, setName] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const nameRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    nameRef.current?.focus();
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onCancel();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onCancel]);

  const trimmed = name.trim();
  const nameValid = trimmed.length > 0 && VALID_NAME.test(trimmed);
  const canSubmit = !!file && nameValid && !submitting;

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!canSubmit) return;
    setSubmitting(true);
    setError(null);
    try {
      await onSubmit(trimmed, file!);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
      setSubmitting(false);
    }
  };

  return (
    <div
      className="fixed inset-0 z-40 flex items-center justify-center p-6 bg-black/60 backdrop-blur-sm animate-fade-in"
      onClick={onCancel}
      role="dialog"
      aria-modal="true"
      aria-label="New version"
    >
      <form
        onSubmit={submit}
        onClick={(e) => e.stopPropagation()}
        className="w-full max-w-[480px] bg-surface border border-border rounded-xl shadow-card flex flex-col"
      >
        <div className="px-6 pt-6 pb-4 border-b border-border">
          <div className="inline-flex items-center gap-2 font-mono text-[11px] tracking-[0.14em] uppercase text-ink-mute mb-3.5">
            <span className="w-1.5 h-1.5 rounded-full bg-accent shadow-[0_0_0_3px_rgba(215,226,122,0.12)]" />
            New version
          </div>
          <h2 className="text-lg font-semibold tracking-[-0.01em] text-ink m-0 mb-1">
            New version of <span className="font-mono">{product.name}</span>
          </h2>
          <p className="text-[13px] text-ink-dim leading-relaxed m-0 mb-4">
            Each version pairs a name (e.g. <span className="font-mono">1.0.0</span>) with one file.
          </p>

          <label className="block mb-3">
            <span className="block font-mono text-[11px] tracking-[0.08em] uppercase text-ink-mute mb-1.5">
              Version name
            </span>
            <input
              ref={nameRef}
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="1.0.0"
              disabled={submitting}
              className="w-full bg-[#0e0e12] border border-border rounded-md px-3 py-2 font-mono text-[13px] text-ink placeholder:text-ink-mute focus:outline-none focus:border-accent focus:ring-[3px] focus:ring-accent/15"
            />
            {trimmed.length > 0 && !nameValid && (
              <span className="mt-1 block text-[11px] text-danger">
                Letters, digits, and . _ + - only.
              </span>
            )}
          </label>

          <label
            htmlFor="version-file"
            className="block cursor-pointer rounded-md border border-dashed border-border-strong bg-[#0e0e12] px-4 py-5 text-center hover:border-accent/60"
          >
            <span className="block font-mono text-[11px] tracking-[0.08em] uppercase text-ink-mute mb-1.5">
              File
            </span>
            <input
              id="version-file"
              type="file"
              className="sr-only"
              disabled={submitting}
              onChange={(e) => setFile(e.target.files?.[0] ?? null)}
            />
            {file ? (
              <span className="block">
                <span className="block text-[13px] text-ink font-medium truncate">{file.name}</span>
                <span className="mt-1 block font-mono text-[11px] text-ink-mute">{formatBytes(file.size)}</span>
              </span>
            ) : (
              <span className="block text-[13px] text-ink-dim">
                Click to pick a file (zip, jar, tar.gz, …)
              </span>
            )}
          </label>
        </div>

        <div className="px-6 py-4 border-t border-border flex items-center justify-between gap-3">
          <div className="min-w-0 text-[12px] text-ink-dim truncate">
            {error ? <span className="text-danger">{error}</span> : <span>Names must be unique within a product.</span>}
          </div>
          <div className="flex gap-2 flex-none">
            <button
              type="button"
              onClick={onCancel}
              disabled={submitting}
              className="bg-transparent text-ink-dim border border-border-strong px-3 py-1.5 rounded-md text-[13px] font-medium whitespace-nowrap hover:bg-[#1b1b22] hover:text-ink active:translate-y-px disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={!canSubmit}
              className="inline-flex items-center gap-2 whitespace-nowrap bg-accent text-accent-ink border border-accent px-3.5 py-1.5 rounded-md text-[13px] font-semibold hover:brightness-105 active:translate-y-px disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {submitting && <Spinner className="!border-accent-ink/30 !border-t-accent-ink" />}
              {submitting ? "Uploading…" : "Upload"}
            </button>
          </div>
        </div>
      </form>
    </div>
  );
}
