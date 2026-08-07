import { useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ChevronDown, Database, Search } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Input } from '@/shared/components/ui/input'

/** What a dataset holds, so the list can offer both levels at once. */
export interface DatasetTypes {
  dataset: string
  dataTypes: string[]
}

interface DataSourceSelectorProps {
  sources: DatasetTypes[]
  dataset: string
  dataType: string | null
  onSelect: (dataset: string, dataType: string | null) => void
}

export function IndexPatternSelector({ sources, dataset, dataType, onSelect }: DataSourceSelectorProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) {
      setQuery('')
      return
    }
    const onDoc = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  // Filtering keeps a dataset whose own name matches, so "alerts" finds the
  // dataset even when none of its types are called that.
  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return sources
    return sources
      .map((s) =>
        s.dataset.toLowerCase().includes(q)
          ? s
          : { ...s, dataTypes: s.dataTypes.filter((d) => d.toLowerCase().includes(q)) },
      )
      .filter((s) => s.dataset.toLowerCase().includes(q) || s.dataTypes.length > 0)
  }, [sources, query])

  // The dataset always shows: "wineventlog" alone does not say whether you are
  // looking at logs or at the alerts raised from them, and both exist.
  const label = `${dataset} · ${dataType ?? t('logExplorer.query.allDataTypes')}`

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen((o) => !o)}
        className={cn(
          'flex h-9 items-center gap-2 rounded-md px-3 text-sm transition-colors',
          open ? 'bg-muted' : 'hover:bg-muted',
        )}
      >
        <Database size={13} className="text-muted-foreground" />
        <span className="font-mono">{label}</span>
        <ChevronDown size={12} className="text-muted-foreground" />
      </button>
      {open && (
        <div className="absolute left-0 top-full z-30 mt-1 flex max-h-80 w-72 flex-col overflow-hidden rounded-md border border-border bg-popover shadow-lg">
          <div className="border-b border-border px-3 py-2 text-[10px] uppercase tracking-wider text-muted-foreground">
            {t('logExplorer.query.dataTypes')}
          </div>
          {sources.length > 0 && (
            <div className="border-b border-border p-2">
              <div className="relative">
                <Search size={13} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
                <Input
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder={t('logExplorer.query.filterDataTypes')}
                  className="h-8 pl-8 text-xs"
                  autoFocus
                />
              </div>
            </div>
          )}
          <div className="min-h-0 flex-1 overflow-y-auto">
            {sources.length === 0 ? (
              <div className="px-3 py-2 text-xs text-muted-foreground">{t('logExplorer.query.noDataTypes')}</div>
            ) : filtered.length === 0 ? (
              <div className="px-3 py-2 text-xs text-muted-foreground">{t('logExplorer.query.noDataTypesMatch')}</div>
            ) : (
              filtered.map((s) => (
                <div key={s.dataset}>
                  {/* The dataset itself is selectable: no data type means every
                      kind of record it holds. */}
                  <button
                    onClick={() => {
                      onSelect(s.dataset, null)
                      setOpen(false)
                    }}
                    className={cn(
                      'flex w-full items-center justify-between gap-2 border-y border-border/60 px-3 py-2 text-left text-xs font-medium uppercase tracking-wider transition-colors',
                      s.dataset === dataset && dataType === null
                        ? 'bg-muted/60 text-foreground'
                        : 'text-muted-foreground hover:bg-muted/60',
                    )}
                  >
                    <span>{s.dataset}</span>
                    <span className="normal-case tracking-normal text-[10px] text-muted-foreground">
                      {t('logExplorer.query.allDataTypes')}
                    </span>
                  </button>
                  {s.dataTypes.map((d) => (
                    <button
                      key={`${s.dataset}:${d}`}
                      onClick={() => {
                        onSelect(s.dataset, d)
                        setOpen(false)
                      }}
                      className={cn(
                        'flex w-full items-center gap-2 py-2 pl-6 pr-3 text-left text-sm transition-colors',
                        s.dataset === dataset && d === dataType ? 'bg-muted/60' : 'hover:bg-muted/60',
                      )}
                    >
                      <span className="font-mono">{d}</span>
                    </button>
                  ))}
                </div>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  )
}
