import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import type { Tab } from '@/features/integrations/types'

interface IntegrationsTabsProps {
  current: Tab
  onChange: (tab: Tab) => void
  counts: Record<Tab, number>
}

const TABS: { id: Tab; labelKey: string }[] = [
  { id: 'all', labelKey: 'integrations.tabs.all' },
  { id: 'agents', labelKey: 'integrations.tabs.agents' },
  { id: 'collectors', labelKey: 'integrations.tabs.collectors' },
  { id: 'cloud', labelKey: 'integrations.tabs.cloud' },
  { id: 'custom', labelKey: 'integrations.tabs.custom' },
]

export function IntegrationsTabs({ current, onChange, counts }: IntegrationsTabsProps) {
  const { t } = useTranslation()

  return (
    <div className="flex flex-wrap items-center gap-1 border-b border-border">
      {TABS.map((tab) => {
        const active = current === tab.id
        return (
          <button
            key={tab.id}
            onClick={() => onChange(tab.id)}
            className={cn(
              'group relative flex items-center gap-2 px-3 py-2 text-xs transition-colors',
              active ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
            )}
          >
            <span>{t(tab.labelKey)}</span>
            <span
              className={cn(
                'rounded-md px-1.5 py-0.5 font-mono text-[10px] tabular-nums',
                active ? 'bg-primary/15 text-primary' : 'bg-muted text-muted-foreground'
              )}
            >
              {counts[tab.id]}
            </span>
            {active && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
          </button>
        )
      })}
    </div>
  )
}
