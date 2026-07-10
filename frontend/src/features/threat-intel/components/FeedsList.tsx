import { useTiFeeds } from '../hooks/use-ti-feeds'
import { FeedRow } from './FeedRow'

const FEED_COLS = '12px 1fr 110px 110px 100px 110px 36px'

export function FeedsList() {
  const { data, isLoading } = useTiFeeds()

  // Handle TiResult union
  if (data?.kind === 'not-configured') return null

  if (isLoading) {
    return (
      <div className="overflow-hidden rounded-xl border border-border bg-card">
        <div
          className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
          style={{ gridTemplateColumns: FEED_COLS }}
        >
          <div />
          <div>Feed</div>
          <div>Kind</div>
          <div className="text-right">Total</div>
          <div className="text-right">+24h</div>
          <div>Last sync</div>
          <div />
        </div>
        {Array.from({ length: 5 }).map((_, i) => (
          <div
            key={i}
            className="h-10 animate-pulse border-b border-border/60 bg-muted/20 px-4"
          />
        ))}
      </div>
    )
  }

  const feeds = data?.kind === 'ok' ? data.value : []

  if (feeds.length === 0) {
    return (
      <div className="rounded-xl border border-border bg-card px-6 py-16 text-center text-sm text-muted-foreground">
        No feeds configured.
      </div>
    )
  }

  return (
    <div className="overflow-hidden rounded-xl border border-border bg-card">
      <div
        className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
        style={{ gridTemplateColumns: FEED_COLS }}
      >
        <div />
        <div>Feed</div>
        <div>Kind</div>
        <div className="text-right">Total</div>
        <div className="text-right">+24h</div>
        <div>Last sync</div>
        <div />
      </div>
      {feeds.map((feed) => (
        <FeedRow key={feed.id} feed={feed} />
      ))}
    </div>
  )
}
