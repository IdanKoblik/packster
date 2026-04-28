function SkeletonRow() {
  return (
    <div
      className="grid grid-cols-[auto_1fr_auto] items-center gap-3.5 px-4 py-3.5 bg-surface-2 border border-border rounded-lg"
      aria-hidden="true"
    >
      <div className="shimmer-bar w-8 h-8 rounded-md" />
      <div>
        <div className="shimmer-bar h-2.5 w-[55%] mb-2" />
        <div className="shimmer-bar h-[9px] w-[75%]" />
      </div>
      <div className="shimmer-bar w-16 h-[22px] rounded-md" />
    </div>
  );
}

export function SkeletonList({ count = 3 }: { count?: number }) {
  return (
    <div className="flex flex-col gap-2">
      {Array.from({ length: count }, (_, i) => (
        <SkeletonRow key={i} />
      ))}
    </div>
  );
}
