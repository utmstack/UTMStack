export function SkeletonRows({ count = 5 }: { count?: number }) {
  return (
    <>
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="h-10 animate-pulse border-b border-border/60 bg-muted/20" />
      ))}
    </>
  )
}
