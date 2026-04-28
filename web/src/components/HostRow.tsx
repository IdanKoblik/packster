import { useState } from "react";
import type { Host } from "../types";
import { parseHost, providerLabel } from "../lib";
import { GithubIcon, GitlabIcon, LockIcon, ArrowIcon, Spinner } from "./icons";

interface Props {
  host: Host;
  disabledByPeer: boolean;
  onSelect: (host: Host) => void;
}

export function HostRow({ host, disabledByPeer, onSelect }: Props) {
  const [redirecting, setRedirecting] = useState(false);
  const parsed = parseHost(host.url);
  const Icon = host.type === "github" ? GithubIcon : GitlabIcon;

  const disabled = redirecting || disabledByPeer;

  const handleClick = () => {
    if (disabled) return;
    setRedirecting(true);
    onSelect(host);
  };

  return (
    <button
      type="button"
      onClick={handleClick}
      disabled={disabled}
      className={[
        "group w-full grid grid-cols-[auto_1fr_auto] items-center gap-3.5",
        "px-4 py-3.5 pl-4",
        "bg-surface-2 border border-border rounded-lg text-left",
        "transition duration-150",
        "hover:bg-[#1b1b22] hover:border-border-strong",
        "focus:outline-none focus-visible:border-accent focus-visible:ring-[3px] focus-visible:ring-accent/15",
        "active:translate-y-px",
        disabled ? "cursor-progress opacity-80" : "cursor-pointer",
      ].join(" ")}
    >
      {/* provider icon tile */}
      <span className="w-8 h-8 rounded-md bg-[#0e0e12] border border-border grid place-items-center text-ink-dim">
        <Icon className="w-4 h-4" />
      </span>

      {/* meta */}
      <span className="min-w-0">
        <span className="block text-sm font-medium text-ink leading-tight">
          Continue with {providerLabel(host.type)}
        </span>
        <span className="mt-1 flex items-center gap-1.5 font-mono text-xs text-ink-mute leading-tight truncate">
          <span className={parsed.secure ? "" : "text-[#b98a5a]"}>
            <LockIcon open={!parsed.secure} className="w-2.5 h-2.5" />
          </span>
          <span className="truncate">{parsed.host}</span>
        </span>
      </span>

      {/* cta */}
      <span
        className={[
          "inline-flex items-center gap-1.5 text-[13px] font-medium whitespace-nowrap",
          "px-2.5 py-1.5 rounded-md border border-transparent",
          "text-ink-dim transition-colors duration-150",
          !disabled &&
            "group-hover:text-ink group-hover:bg-[#22222a] group-hover:border-border-strong",
        ]
          .filter(Boolean)
          .join(" ")}
      >
        {redirecting ? (
          <>
            <Spinner />
            <span>Redirecting…</span>
          </>
        ) : (
          <>
            <span>Sign in</span>
            <ArrowIcon className="transition-transform duration-200 group-hover:translate-x-0.5" />
          </>
        )}
      </span>
    </button>
  );
}
