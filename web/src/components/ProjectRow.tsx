import type { Project } from "../types";
import { ArrowIcon, GitlabIcon, GithubIcon } from "./icons";
import type { HostType } from "../types";

interface Props {
  project: Project;
  hostType: HostType;
  onOpen: () => void;
}

export function ProjectRow({ project, hostType, onOpen }: Props) {
  const Icon = hostType === "github" ? GithubIcon : GitlabIcon;
  const path = project.web_url.replace(/^https?:\/\//, "");
  return (
    <button
      type="button"
      onClick={onOpen}
      className="group grid grid-cols-[auto_1fr_auto_auto] items-center gap-3 bg-surface-2 border border-border rounded-lg px-3.5 py-3 transition duration-150 text-left hover:bg-[#1b1b22] hover:border-border-strong focus:outline-none focus-visible:border-accent focus-visible:ring-[3px] focus-visible:ring-accent/15 active:translate-y-px"
    >
      <span className="w-8 h-8 rounded-md bg-[#0e0e12] border border-border grid place-items-center text-ink-dim">
        <Icon className="w-4 h-4" />
      </span>
      <span className="min-w-0">
        <span className="block text-sm font-medium text-ink truncate leading-tight">{project.name}</span>
        <span className="mt-1 block font-mono text-xs text-ink-mute truncate leading-tight">{path}</span>
      </span>
      <span className="font-mono text-[10px] tracking-[0.08em] uppercase text-ink-mute px-1.5 py-0.5 border border-border rounded bg-[#0e0e12] whitespace-nowrap">
        org #{project.org}
      </span>
      <ArrowIcon className="text-ink-mute transition-transform duration-200 group-hover:translate-x-0.5 group-hover:text-ink-dim" />
    </button>
  );
}
