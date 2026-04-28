import type { HealthState } from "../types";

interface Props {
  state: HealthState;
  label: string;
}

export function HealthIndicator({ state, label }: Props) {
  const dotColor =
    state === "ok"
      ? "bg-ok shadow-[0_0_0_3px_rgba(107,208,122,0.14)]"
      : state === "down"
      ? "bg-danger shadow-[0_0_0_3px_rgba(239,106,106,0.14)]"
      : "bg-ink-mute shadow-[0_0_0_3px_rgba(255,255,255,0.04)]";

  return (
    <div className="inline-flex items-center gap-1.5 font-mono text-[11px] tracking-[0.08em] uppercase">
      <span className={`w-1.5 h-1.5 rounded-full transition duration-200 ${dotColor}`} />
      <span>{label}</span>
    </div>
  );
}
