import { ChevronDown, Flame, Tag as TagIcon } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { STATUS_INT, type AlertTag, type StatusKey } from '../types/alert.types'
import { Menu } from './menu'

export function AlertsBulkBar({
  count,
  tagCatalog,
  onClear,
  onStatus,
  onTag,
  onIncident,
}: {
  count: number
  tagCatalog: AlertTag[]
  onClear: () => void
  onStatus: (status: number) => void
  onTag: (tag: string) => void
  onIncident: () => void
}) {
  const { t } = useTranslation()
  return (
    <div className="mt-3 flex flex-wrap items-center gap-2 rounded-lg border border-primary/30 bg-primary/5 px-3 py-2 text-xs">
      <span className="font-medium">{t('alerts.bulk.selected', { count })}</span>
      <span className="text-border">·</span>
      <Menu trigger={<>{t('alerts.bulk.setStatus')} <ChevronDown size={12} /></>}>
        {(['open', 'in_review', 'completed'] as StatusKey[]).map((k) => (
          <button
            key={k}
            onClick={() => onStatus(STATUS_INT[k])}
            className="block w-full px-3 py-1.5 text-left text-sm hover:bg-muted"
          >
            {t(`alerts.status.${k}`)}
          </button>
        ))}
      </Menu>
      <Menu trigger={<><TagIcon size={12} /> {t('alerts.bulk.addTag')} <ChevronDown size={12} /></>}>
        {tagCatalog.length === 0 && (
          <div className="px-3 py-1.5 text-xs text-muted-foreground">{t('alerts.bulk.noTags')}</div>
        )}
        {tagCatalog.map((tg) => (
          <button
            key={tg.id}
            onClick={() => onTag(tg.tagName)}
            className="block w-full px-3 py-1.5 text-left text-sm hover:bg-muted"
          >
            {tg.tagName}
          </button>
        ))}
      </Menu>
      <button
        onClick={onIncident}
        className="inline-flex items-center gap-1.5 rounded-md border border-border bg-card px-2 py-1 hover:bg-muted"
      >
        <Flame size={12} className="text-red-500" /> {t('alerts.bulk.addToIncident')}
      </button>
      <button onClick={onClear} className="ml-auto text-muted-foreground hover:text-foreground">
        {t('alerts.bulk.clear')}
      </button>
    </div>
  )
}
