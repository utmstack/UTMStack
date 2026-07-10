import { cn } from '@/shared/lib/utils'
import type { ThreatFeed } from '../domain/threat-intel.types'
import { feedAccuracyMeta, feedTypeTone } from './utils/severity-style'

interface FeedRowProps {
  feed: ThreatFeed
}

const FEED_COLS = '12px 1fr 160px 140px'

export function FeedRow({ feed }: FeedRowProps) {
  const acc = feedAccuracyMeta(feed.accuracy)

  return (
    <div
      className="group grid items-center gap-3 border-b border-border/60 px-4 py-2.5 text-xs last:border-b-0 hover:bg-muted/40"
      style={{ gridTemplateColumns: FEED_COLS }}
    >
      <span className={cn('h-2 w-2 rounded-full', acc.dot)} title={acc.label} />
      <div className="min-w-0 truncate font-medium">{feed.name}</div>
      <div className={cn('text-[11px] capitalize', feedTypeTone(feed.type))}>{feed.type}</div>
      <div className="text-[11px] text-muted-foreground">{acc.label}</div>
    </div>
  )
}
