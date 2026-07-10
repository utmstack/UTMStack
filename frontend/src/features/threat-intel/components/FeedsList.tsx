import { useTranslation } from 'react-i18next'
import { useTiFeeds } from '../hooks/use-ti-feeds'
import { FeedRow } from './FeedRow'
import { FeedsHeader } from './FeedsHeader'

export function FeedsList() {
  const { t } = useTranslation()
  const { data, isLoading } = useTiFeeds()

  if (data?.kind === 'not-configured') return null

  if (isLoading) {
    return (
      <div className="overflow-hidden rounded-xl border border-border bg-card">
        <FeedsHeader />
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
    <div className="overflow-hidden overflow-y-auto max-h-[70dvh] rounded-xl border border-border bg-card">
      <FeedsHeader />
      {feeds.map((feed) => (
        <FeedRow key={`${feed.type}-${feed.name}`} feed={feed} />
      ))}
    </div>
  )
}
