import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { STATUS_INT, STATUS_TABS, type StatusKey, type StatusTab } from '../types/alert.types'

export function AlertsStatusTabs({
  current,
  counts,
  onChange,
}: {
  current: StatusTab
  counts: Record<string, number>
  onChange: (s: StatusTab) => void
}) {
  const { t } = useTranslation()
  const label = (id: StatusTab) => (id === 'all' ? t('alerts.statusTabs.all') : t(`alerts.status.${id}`))
  const count = (id: StatusTab) => (id === 'all' ? undefined : counts[String(STATUS_INT[id as StatusKey])])
  return (
    <div className="flex flex-wrap items-center gap-1 border-b border-border">
      {STATUS_TABS.map((id) => {
        const active = current === id
        const c = count(id)
        return (
          <button
            key={id}
            onClick={() => onChange(id)}
            className={cn(
              'relative flex items-center gap-2 px-3 py-2 text-xs transition-colors',
              active ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
            )}
          >
            {label(id)}
            {c != null && (
              <span
                className={cn(
                  'rounded-md px-1.5 py-0.5 font-mono text-[10px] tabular-nums',
                  active ? 'bg-primary/15 text-primary' : 'bg-muted text-muted-foreground'
                )}
              >
                {c}
              </span>
            )}
            {active && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
          </button>
        )
      })}
    </div>
  )
}
