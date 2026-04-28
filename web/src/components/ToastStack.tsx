import type { Toast } from "../types";

interface Props {
  toasts: Toast[];
  onDismiss: (id: number) => void;
}

export function ToastStack({ toasts, onDismiss }: Props) {
  if (!toasts.length) return null;
  return (
    <div className="fixed top-5 right-5 z-50 flex flex-col gap-2">
      {toasts.map((t) => (
        <div
          key={t.id}
          role="status"
          className={[
            "min-w-[240px] max-w-[320px] px-3.5 py-2.5 rounded-lg border text-[13px] shadow-card animate-fade-in flex items-start gap-3",
            t.kind === "error"
              ? "bg-[#2a1619] border-[#3a2326] text-[#ffc9c9]"
              : t.kind === "success"
              ? "bg-[#18251c] border-[#25382a] text-[#c6e7cd]"
              : "bg-surface border-border text-ink",
          ].join(" ")}
        >
          <span className="flex-1 leading-snug">{t.message}</span>
          <button
            onClick={() => onDismiss(t.id)}
            className="text-ink-mute hover:text-ink text-xs leading-none pt-0.5"
            aria-label="Dismiss"
          >
            ✕
          </button>
        </div>
      ))}
    </div>
  );
}
