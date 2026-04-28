import type { SVGProps } from "react";

type Props = SVGProps<SVGSVGElement>;

export function GithubIcon(props: Props) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true" {...props}>
      <path d="M12 .5C5.73.5.75 5.48.75 11.75c0 4.97 3.22 9.18 7.69 10.67.56.1.77-.24.77-.54 0-.27-.01-.97-.02-1.9-3.13.68-3.79-1.51-3.79-1.51-.51-1.3-1.25-1.64-1.25-1.64-1.02-.7.08-.69.08-.69 1.13.08 1.72 1.16 1.72 1.16 1 1.72 2.63 1.22 3.27.93.1-.73.39-1.22.71-1.5-2.5-.28-5.13-1.25-5.13-5.57 0-1.23.44-2.24 1.16-3.03-.12-.29-.5-1.44.11-3 0 0 .95-.3 3.11 1.16.9-.25 1.87-.38 2.83-.38.96 0 1.93.13 2.83.38 2.15-1.46 3.1-1.16 3.1-1.16.62 1.56.23 2.71.12 3 .72.79 1.16 1.8 1.16 3.03 0 4.33-2.64 5.28-5.15 5.56.4.34.76 1.03.76 2.08 0 1.5-.01 2.71-.01 3.08 0 .3.2.65.78.54 4.46-1.49 7.68-5.7 7.68-10.67C23.25 5.48 18.27.5 12 .5z" />
    </svg>
  );
}

export function GitlabIcon(props: Props) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true" {...props}>
      <path d="M23.955 13.587l-1.342-4.135-2.664-8.189a.456.456 0 00-.867 0L16.418 9.45H7.582L4.918 1.263a.456.456 0 00-.867 0L1.387 9.452.045 13.587a.924.924 0 00.331 1.023L12 23.054l11.625-8.444a.924.924 0 00.33-1.023" />
    </svg>
  );
}

export function LockIcon({ open = false, ...props }: Props & { open?: boolean }) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" aria-hidden="true" {...props}>
      <rect x="4" y="11" width="16" height="10" rx="2" />
      {open ? <path d="M8 11V8a4 4 0 0 1 7.5-2" /> : <path d="M8 11V8a4 4 0 0 1 8 0v3" />}
    </svg>
  );
}

export function ArrowIcon(props: Props) {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" aria-hidden="true" {...props}>
      <path d="M5 12h14M13 6l6 6-6 6" />
    </svg>
  );
}

export function EmptyIcon(props: Props) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.6} strokeLinecap="round" strokeLinejoin="round" aria-hidden="true" {...props}>
      <path d="M3 7l9-4 9 4M3 7v10l9 4 9-4V7M3 7l9 4 9-4M12 11v10" />
    </svg>
  );
}

export function AlertIcon(props: Props) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.6} strokeLinecap="round" strokeLinejoin="round" aria-hidden="true" {...props}>
      <path d="M12 9v4M12 17h.01M10.29 3.86l-8.19 14A2 2 0 0 0 3.83 21h16.34a2 2 0 0 0 1.73-3.14l-8.19-14a2 2 0 0 0-3.46 0z" />
    </svg>
  );
}

export function RefreshIcon(props: Props) {
  return (
    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" aria-hidden="true" {...props}>
      <path d="M21 12a9 9 0 1 1-3-6.7" />
      <path d="M21 3v6h-6" />
    </svg>
  );
}

export function PlusIcon(props: Props) {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.5} strokeLinecap="round" aria-hidden="true" {...props}>
      <path d="M12 5v14M5 12h14" />
    </svg>
  );
}

export function Spinner({ className = "" }: { className?: string }) {
  return (
    <span
      className={`inline-block w-3 h-3 rounded-full border-[1.5px] border-white/10 border-t-ink animate-spin ${className}`}
      aria-hidden="true"
    />
  );
}
