const FEED_COLS = '12px 1fr 160px 140px'

export function FeedsHeader() {
  return (
    <div
      className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
      style={{ gridTemplateColumns: FEED_COLS }}
    >
      <div />
      <div>Feed</div>
      <div>Type</div>
      <div>Accuracy</div>
    </div>
  )
}
