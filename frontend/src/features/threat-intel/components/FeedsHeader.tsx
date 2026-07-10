import { useTranslation } from 'react-i18next'

const FEED_COLS = '12px 1fr 160px 140px'

export function FeedsHeader() {
  const { t } = useTranslation()
  return (
    <div
      className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
      style={{ gridTemplateColumns: FEED_COLS }}
    >
      <div />
      <div>{t('threatIntel.feeds.table.name')}</div>
      <div>{t('threatIntel.feeds.table.type')}</div>
      <div>{t('threatIntel.feeds.table.accuracy')}</div>
    </div>
  )
}
