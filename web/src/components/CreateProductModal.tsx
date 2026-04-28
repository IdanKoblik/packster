import { useEffect, useRef, useState } from "react";
import { Spinner } from "./icons";

interface Props {
  onCancel: () => void;
  onSubmit: (name: string) => Promise<void>;
}

export function CreateProductModal({ onCancel, onSubmit }: Props) {
  const [name, setName] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    inputRef.current?.focus();
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onCancel();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onCancel]);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = name.trim();
    if (!trimmed || submitting) return;
    setSubmitting(true);
    setError(null);
    try {
      await onSubmit(trimmed);
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
      aria-label="Create product"
    >
      <form
        onSubmit={submit}
        onClick={(e) => e.stopPropagation()}
        className="w-full max-w-[440px] bg-surface border border-border rounded-xl shadow-card flex flex-col"
      >
        <div className="px-6 pt-6 pb-4 border-b border-border">
          <div className="inline-flex items-center gap-2 font-mono text-[11px] tracking-[0.14em] uppercase text-ink-mute mb-3.5">
            <span className="w-1.5 h-1.5 rounded-full bg-accent shadow-[0_0_0_3px_rgba(215,226,122,0.12)]" />
            New product
          </div>
          <h2 className="text-lg font-semibold tracking-[-0.01em] text-ink m-0 mb-1">Add a product</h2>
          <p className="text-[13px] text-ink-dim leading-relaxed m-0 mb-4">
            A product groups versioned uploads under a project (e.g. <span className="font-mono">spigot</span>).
          </p>
          <input
            ref={inputRef}
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="product name"
            disabled={submitting}
            className="w-full bg-[#0e0e12] border border-border rounded-md px-3 py-2 text-[13px] text-ink placeholder:text-ink-mute focus:outline-none focus:border-accent focus:ring-[3px] focus:ring-accent/15"
          />
        </div>

        <div className="px-6 py-4 border-t border-border flex items-center justify-between gap-3">
          <div className="min-w-0 text-[12px] text-ink-dim truncate">
            {error ? <span className="text-danger">{error}</span> : <span>Names must be unique within a project.</span>}
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
              disabled={!name.trim() || submitting}
              className="inline-flex items-center gap-2 whitespace-nowrap bg-accent text-accent-ink border border-accent px-3.5 py-1.5 rounded-md text-[13px] font-semibold hover:brightness-105 active:translate-y-px disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {submitting && <Spinner className="!border-accent-ink/30 !border-t-accent-ink" />}
              {submitting ? "Creating…" : "Create product"}
            </button>
          </div>
        </div>
      </form>
    </div>
  );
}
