import type { DebugState } from "../types";

interface Props {
  state: DebugState;
  onChange: (s: DebugState) => void;
}

const STATES: DebugState[] = ["success", "loading", "empty", "error"];

export function DebugStateSwitch({ state, onChange }: Props) {
  return (
    <div
      role="tablist"
      aria-label="Preview state"
      className="fixed bottom-5 right-5 z-50 flex gap-0.5 p-1.5 bg-surface border border-border rounded-lg font-mono text-[11px] shadow-[0_20px_40px_-20px_rgba(0,0,0,0.8)]"
    >
      <span className="absolute -top-[18px] left-0 text-[9px] tracking-[0.16em] text-ink-mute">
        STATE
      </span>
      {STATES.map((s) => (
        <button
          key={s}
          role="tab"
          aria-selected={state === s}
          onClick={() => onChange(s)}
          className={[
            "px-2.5 py-1.5 rounded-md uppercase tracking-[0.08em] border-none transition-colors",
            state === s
              ? "bg-accent text-accent-ink"
              : "text-ink-dim hover:text-ink hover:bg-white/5",
          ].join(" ")}
        >
          {s}
        </button>
      ))}
    </div>
  );
}
