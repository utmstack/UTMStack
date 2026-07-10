import { cn } from '@/shared/lib/utils'

export type TabKey = 'iocs' | 'actors' | 'feeds'

export interface TabsRowProps {
  active: TabKey
  onChange: (t: TabKey) => void
  counts?: { iocs?: number; actors?: number; feeds?: number }
}

const TABS: { id: TabKey; label: string }[] = [
  { id: 'iocs', label: 'IOCs' },
  { id: 'actors', label: 'Threat actors' },
  { id: 'feeds', label: 'Feeds' },
]

export function TabsRow({ active, onChange, counts }: TabsRowProps) {
  return (
    <div className="flex flex-wrap items-center gap-1 border-b border-border">
      {TABS.map((t) => {
        const tabActive = active === t.id
        const count = counts?.[t.id] ?? 0
        return (
          <button
            key={t.id}
            onClick={() => onChange(t.id)}
            className={cn(
              'group relative flex items-center gap-2 px-3 py-2 text-xs transition-colors',
              tabActive ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
            )}
          >
            <span>{t.label}</span>
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
