import { useMemo, useState } from 'react'
import { AlertTriangle, LayoutGrid, ListIcon, Loader2, RefreshCw, Search, ShieldAlert } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { VIEWS, type ViewId } from '../lib/adversary-meta'
import { useAdversaries } from '../hooks/use-adversaries'
import type { Adversary } from '../types/adversary.types'
import { AdversariesOverviewCard } from '../components/adversaries-overview-card'
import { AdversariesList } from '../components/adversaries-list'
import { AdversariesGrid } from '../components/adversaries-grid'
import { AdversaryDrawer } from '../components/adversary-drawer'

export function AdversariesPage() {
  const { t } = useTranslation()
  const [view, setView] = useState<ViewId>('all')
  const [search, setSearch] = useState('')
  const [layout, setLayout] = useState<'list' | 'cards'>('list')
  const [open, setOpen] = useState<Adversary | null>(null)
  const { adversaries, loading, error, refresh } = useAdversaries()

  const counts = useMemo(
    () => Object.fromEntries(VIEWS.map((v) => [v.id, adversaries.filter(v.predicate).length])),
    [adversaries],
  )

  const filtered = useMemo(() => {
    const v = VIEWS.find((x) => x.id === view) ?? VIEWS[0]
    const q = search.trim().toLowerCase()
    return adversaries.filter(v.predicate).filter((a) =>
      q
        ? (a.identifier + ' ' + (a.geo?.country ?? '') + ' ' + (a.geo?.aso ?? '')).toLowerCase().includes(q)
        : true,
    )
  }, [adversaries, view, search])

  return (
    <div className="mx-auto w-full max-w-[1100px] px-6 py-6">
      <header className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <div className="flex items-center gap-3">
            <h1 className="flex items-center gap-2 text-xl font-semibold">
              <ShieldAlert size={18} strokeWidth={1.75} />
              {t('adversaries.title')}
            </h1>
            <span className="rounded-md bg-muted px-2 py-0.5 font-mono text-[11px] tabular-nums text-muted-foreground">
              {t('adversaries.countChip', { count: adversaries.length })}
            </span>
          </div>
          <p className="mt-1 text-xs text-muted-foreground">{t('adversaries.subtitle')}</p>
        </div>
        <button
          onClick={refresh}
          title={t('adversaries.refresh')}
          className="flex h-9 w-9 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
        </button>
      </header>

      <div className="mt-5">
        <AdversariesOverviewCard adversaries={adversaries} />
      </div>

      <div className="mt-5 flex flex-wrap items-center gap-1 border-b border-border">
        {VIEWS.map((v) => {
          const active = view === v.id
          return (
            <button
              key={v.id}
              onClick={() => setView(v.id)}
              className={cn(
                'relative flex items-center gap-2 px-3 py-2 text-xs transition-colors',
                active ? 'text-foreground' : 'text-muted-foreground hover:text-foreground',
              )}
            >
              <span>{t(`adversaries.views.${v.id}`)}</span>
              <span
                className={cn(
                  'rounded-md px-1.5 py-0.5 font-mono text-[10px] tabular-nums',
                  active ? 'bg-primary/15 text-primary' : 'bg-muted text-muted-foreground',
                )}
              >
                {counts[v.id] ?? 0}
              </span>
              {active && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
            </button>
          )
        })}
      </div>

      <div className="mt-3 flex flex-wrap items-center gap-2">
        <div className="relative min-w-[260px] flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t('adversaries.search')}
            className="h-9 pl-9"
          />
        </div>
        <div className="ml-auto flex items-center gap-1 rounded-md border border-border p-0.5">
          <button
            onClick={() => setLayout('list')}
            title={t('adversaries.layout.list')}
            className={cn(
              'flex h-7 w-7 items-center justify-center rounded transition-colors',
              layout === 'list' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground',
            )}
          >
            <ListIcon size={13} />
          </button>
          <button
            onClick={() => setLayout('cards')}
            title={t('adversaries.layout.cards')}
            className={cn(
              'flex h-7 w-7 items-center justify-center rounded transition-colors',
              layout === 'cards' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground',
            )}
          >
            <LayoutGrid size={13} />
          </button>
        </div>
      </div>

      {error ? (
        <div className="mt-3 flex flex-col items-center gap-3 rounded-xl border border-border bg-card px-6 py-12 text-sm">
          <span className="inline-flex items-center gap-2 text-muted-foreground">
            <AlertTriangle size={16} className="text-amber-500" />
            {t('adversaries.loadError')}
          </span>
          <Button variant="outline" size="sm" onClick={refresh}>
            {t('adversaries.retry')}
          </Button>
        </div>
      ) : loading ? (
        <div className="mt-3 flex items-center justify-center gap-2 rounded-xl border border-border bg-card px-6 py-16 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" />
        </div>
      ) : filtered.length === 0 ? (
        <div className="mt-3 rounded-xl border border-border bg-card px-6 py-16 text-center text-sm text-muted-foreground">
          {t('adversaries.empty')}
        </div>
      ) : layout === 'list' ? (
        <AdversariesList adversaries={filtered} onOpen={setOpen} />
      ) : (
        <AdversariesGrid adversaries={filtered} onOpen={setOpen} />
      )}

      {open && <AdversaryDrawer a={open} onClose={() => setOpen(null)} />}
    </div>
  )
}
