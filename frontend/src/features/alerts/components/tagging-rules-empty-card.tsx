export function TaggingRulesEmptyCard({ children }: { children: React.ReactNode }) {
  return (
    <div className="mt-4 flex flex-1 items-center justify-center gap-2 rounded-xl border border-border bg-card text-sm text-muted-foreground">
      {children}
    </div>
  )
}
