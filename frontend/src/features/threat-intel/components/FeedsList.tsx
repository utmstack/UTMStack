import { useTranslation } from 'react-i18next'
import { colMins, useResizableColumns } from '@/shared/hooks/useResizableColumns'
import { useTiFeeds } from '../hooks/use-ti-feeds'
import { FeedRow } from './FeedRow'
import { FeedsHeader } from './FeedsHeader'

const FEED_COLS = [12, '1fr', 160, 140]

export function FeedsList() {
  const { t } = useTranslation()
  const { data, isLoading } = useTiFeeds()
  const feedLabelMins = colMins([
    t('threatIntel.feeds.table.name'),
    t('threatIntel.feeds.table.type'),
    t('threatIntel.feeds.table.accuracy'),
  ])
  const { template: tableCols, startDrag } = useResizableColumns(FEED_COLS, {
    min: [12, ...feedLabelMins],
    storageKey: 'threat-intel-feed-table-columns',
  })

  if (data?.kind === 'not-configured') return null

  if (isLoading) {
    return (
      <div className="overflow-x-auto overflow-y-hidden rounded-xl border border-border bg-card">
        <FeedsHeader tableCols={tableCols} startDrag={startDrag} />
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
        {t('threatIntel.feeds.empty')}
      </div>
    )
  }

  return (
    <div className="max-h-[70dvh] overflow-x-auto overflow-y-auto rounded-xl border border-border bg-card">
      <FeedsHeader tableCols={tableCols} startDrag={startDrag} />
      {feeds.map((feed) => (
        <FeedRow key={`${feed.type}-${feed.name}`} feed={feed} tableCols={tableCols} />
      ))}
    </div>
  )
}
