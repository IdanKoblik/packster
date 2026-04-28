import { useState } from "react";
import type { Permission, Version } from "../types";
import { Spinner } from "./icons";

interface Props {
  version: Version;
  permission: Permission;
  onDownload: () => Promise<void>;
  onDelete: () => Promise<void>;
}

export function VersionRow({ version, permission, onDownload, onDelete }: Props) {
  const [downloading, setDownloading] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [copied, setCopied] = useState(false);

  const trimmedChecksum = version.checksum.slice(0, 12);

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(version.checksum);
      setCopied(true);
      setTimeout(() => setCopied(false), 1200);
    } catch {
      /* clipboard unavailable */
    }
  };

  const doDownload = async () => {
    if (downloading) return;
    setDownloading(true);
    try {
      await onDownload();
    } finally {
      setDownloading(false);
    }
  };

  const doDelete = async () => {
    if (deleting) return;
    if (!confirm(`Delete version ${version.name}? This cannot be undone.`)) return;
    setDeleting(true);
    try {
      await onDelete();
    } finally {
      setDeleting(false);
    }
  };

  return (
    <div className="grid grid-cols-[1fr_auto] items-center gap-3 px-3 py-2 rounded-md hover:bg-[#1b1b22]">
      <div className="min-w-0">
        <div className="flex items-center gap-2 min-w-0">
          <span className="font-mono text-[13px] text-ink font-semibold truncate">{version.name}</span>
          <span className="font-mono text-[11px] text-ink-mute truncate">{version.path}</span>
        </div>
        <button
          type="button"
          onClick={copy}
          className="mt-0.5 inline-flex items-center gap-1.5 font-mono text-[11px] text-ink-mute hover:text-ink"
          title={version.checksum}
        >
          <span>sha256:{trimmedChecksum}…</span>
          <span className="text-[10px] opacity-60">{copied ? "copied" : "copy"}</span>
        </button>
      </div>
      <div className="flex gap-1.5">
        {permission.can_download && (
          <button
            type="button"
            onClick={doDownload}
            disabled={downloading}
            className="inline-flex items-center gap-1.5 bg-transparent text-ink border border-border-strong px-2.5 py-1 rounded-md text-[12px] font-medium hover:bg-[#1b1b22] hover:border-[#3a3a44] active:translate-y-px disabled:opacity-50"
          >
            {downloading && <Spinner className="!w-2.5 !h-2.5" />}
            {downloading ? "Downloading…" : "Download"}
          </button>
        )}
        {permission.can_delete && (
          <button
            type="button"
            onClick={doDelete}
            disabled={deleting}
            className="inline-flex items-center gap-1.5 bg-transparent text-danger border border-[#3a2326] px-2.5 py-1 rounded-md text-[12px] font-medium hover:bg-[#2a1619] active:translate-y-px disabled:opacity-50"
          >
            {deleting ? "Deleting…" : "Delete"}
          </button>
        )}
      </div>
    </div>
  );
}
