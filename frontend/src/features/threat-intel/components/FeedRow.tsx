import { MoreHorizontal } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import type { ThreatFeed } from '../domain/threat-intel.types'
import { FEED_STATUS_META, FEED_KIND_TONE } from '../components/utils/severity-style'
import { relativeTime } from '../components/utils/time-format'

interface FeedRowProps {
  feed: ThreatFeed
}

const FEED_COLS = '12px 1fr 110px 110px 100px 110px 36px'

export function FeedRow({ feed }: FeedRowProps) {
  const st = FEED_STATUS_META[feed.status]

  return (
    <div
      className="group grid cursor-pointer items-center gap-3 border-b border-border/60 px-4 py-2.5 text-xs hover:bg-muted/40 last:border-b-0"
      style={{ gridTemplateColumns: FEED_COLS }}
    >
      <span className={cn('h-2 w-2 rounded-full', st.dot)} title={st.label} />
      <div className="min-w-0">
        <div className="truncate font-medium">{feed.name}</div>
        <div className="truncate text-[10px] text-muted-foreground">
          syncs every{' '}
          {feed.syncIntervalMin < 60
            ? `${feed.syncIntervalMin}m`
            : feed.syncIntervalMin < 1440
              ? `${Math.round(feed.syncIntervalMin / 60)}h`
              : `${Math.round(feed.syncIntervalMin / 1440)}d`}
        </div>
      </div>
      <div className={cn('text-[11px] capitalize', FEED_KIND_TONE[feed.kind])}>
        {feed.kind}
      </div>
      <div className="text-right font-mono tabular-nums text-muted-foreground">
        {feed.itemsTotal.toLocaleString()}
      </div>
      <div
        className={cn(
          'text-right font-mono tabular-nums',
          feed.itemsAdded24h > 0 ? 'text-emerald-500' : 'text-muted-foreground'
        )}
      >
        {feed.itemsAdded24h > 0 ? `+${feed.itemsAdded24h.toLocaleString()}` : '—'}
      </div>
      <div className="font-mono text-[11px] text-muted-foreground">
        {relativeTime(feed.lastSync)}
      </div>
      <div className="flex justify-end opacity-0 group-hover:opacity-100">
        <button className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-background hover:text-foreground">
          <MoreHorizontal size={14} />
        </button>
      </div>
    </div>
  )
}
