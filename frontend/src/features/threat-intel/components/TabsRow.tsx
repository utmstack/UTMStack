import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'

export type TabKey = 'iocs' | 'actors' | 'feeds'

export interface TabsRowProps {
  active: TabKey
  onChange: (t: TabKey) => void
  counts?: { iocs?: number; actors?: number; feeds?: number }
}

export function TabsRow({ active, onChange, counts }: TabsRowProps) {
  const { t } = useTranslation()
  const TABS: { id: TabKey; labelKey: string }[] = [
    { id: 'iocs', labelKey: 'threatIntel.tabs.iocs' },
    { id: 'actors', labelKey: 'threatIntel.tabs.actors' },
    { id: 'feeds', labelKey: 'threatIntel.tabs.feeds' },
  ]
  return (
    <div className="flex flex-wrap items-center gap-1 border-b border-border">
      {TABS.map((tab) => {
        const tabActive = active === tab.id
        const count = counts?.[tab.id] ?? 0
        return (
          <button
            key={tab.id}
            onClick={() => onChange(tab.id)}
            className={cn(
              'group relative flex items-center gap-2 px-3 py-2 text-xs transition-colors',
              tabActive ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
            )}
          >
            <span>{t(tab.labelKey)}</span>
            <span
              className={cn(
                'rounded-md px-1.5 py-0.5 font-mono text-[10px] tabular-nums',
                tabActive ? 'bg-primary/15 text-primary' : 'bg-muted text-muted-foreground'
              )}
            >
              {count}
            </span>
            {tabActive && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
          </button>
        )
      })}
    </div>
  )
}
