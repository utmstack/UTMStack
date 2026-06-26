import { useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ChevronDown, Database, Search } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Input } from '@/shared/components/ui/input'
import type { IndexPattern } from '../types/log-explorer.types'

interface IndexPatternSelectorProps {
  patterns: IndexPattern[]
  pattern: IndexPattern | null
  onPattern: (p: IndexPattern) => void
}

export function IndexPatternSelector({ patterns, pattern, onPattern }: IndexPatternSelectorProps) {
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

  const filtered = useMemo(
    () =>
      query.trim()
        ? patterns.filter((p) => p.pattern.toLowerCase().includes(query.trim().toLowerCase()))
        : patterns,
    [patterns, query]
  )

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen((o) => !o)}
        className={cn(
          'flex h-9 items-center gap-2 rounded-md px-3 text-sm transition-colors',
          open ? 'bg-muted' : 'hover:bg-muted'
        )}
      >
        <Database size={13} className="text-muted-foreground" />
        <span className="font-mono">{pattern?.pattern ?? '—'}</span>
        <ChevronDown size={12} className="text-muted-foreground" />
      </button>
      {open && (
        <div className="absolute left-0 top-full z-30 mt-1 flex max-h-80 w-72 flex-col overflow-hidden rounded-md border border-border bg-popover shadow-lg">
          <div className="border-b border-border px-3 py-2 text-[10px] uppercase tracking-wider text-muted-foreground">
            {t('logExplorer.query.indexPatterns')}
          </div>
          {patterns.length > 0 && (
            <div className="border-b border-border p-2">
              <div className="relative">
                <Search size={13} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
                <Input
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder={t('logExplorer.query.filterPatterns')}
                  className="h-8 pl-8 text-xs"
                  autoFocus
                />
              </div>
            </div>
          )}
          <div className="min-h-0 flex-1 overflow-y-auto">
            {patterns.length === 0 ? (
              <div className="px-3 py-2 text-xs text-muted-foreground">{t('logExplorer.query.noPatterns')}</div>
            ) : filtered.length === 0 ? (
              <div className="px-3 py-2 text-xs text-muted-foreground">{t('logExplorer.query.noPatternsMatch')}</div>
            ) : (
              filtered.map((p) => (
                <button
                  key={p.id}
                  onClick={() => {
                    onPattern(p)
                    setOpen(false)
                  }}
                  className={cn(
                    'flex w-full items-center gap-2 px-3 py-2 text-left text-sm transition-colors',
                    p.id === pattern?.id ? 'bg-muted/60' : 'hover:bg-muted/60'
                  )}
                >
                  <span className="font-mono">{p.pattern}</span>
                </button>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  )
}
