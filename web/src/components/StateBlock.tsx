import type { ReactNode } from "react";
import { RefreshIcon } from "./icons";

interface Props {
  icon: ReactNode;
  iconTone?: "default" | "danger";
  title: string;
  children: ReactNode;
  actionLabel?: string;
  onAction?: () => void;
}

export function StateBlock({
  icon,
  iconTone = "default",
  title,
  children,
  actionLabel,
  onAction,
}: Props) {
  return (
    <div className="rounded-lg border border-dashed border-border-strong bg-white/[0.01] px-5 py-7 text-center">
      <div
        className={[
          "mx-auto mb-3 w-9 h-9 rounded-lg border bg-surface-2 grid place-items-center",
          iconTone === "danger" ? "text-danger border-[#3a2326]" : "text-ink-mute border-border",
        ].join(" ")}
      >
        <span className="w-[18px] h-[18px] block">{icon}</span>
      </div>
      <p className="text-sm font-medium text-ink mb-1">{title}</p>
      <p className="text-[13px] text-ink-dim mb-3.5 leading-relaxed">{children}</p>
      {actionLabel && onAction && (
        <button
          type="button"
          onClick={onAction}
          className="inline-flex items-center gap-2 bg-transparent text-ink border border-border-strong px-3 py-1.5 rounded-md text-[13px] font-medium transition-colors hover:bg-[#1b1b22] hover:border-[#3a3a44] active:translate-y-px"
        >
          <RefreshIcon />
          {actionLabel}
        </button>
      )}
    </div>
  );
}
