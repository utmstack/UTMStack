import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, Loader2, Lock, Plus, RefreshCw, Search, ShieldCheck } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { complianceService } from '../services/compliance-http.service'
import type { Control } from '../types/compliance.types'
import { copyOfControl } from '../lib/duplicate'
import { StartFromModal } from './StartFromModal'
import { ControlEditor } from './ControlEditor'

// Same rule as the frameworks grid: the column count follows the card's own
// width, so a wide monitor adds columns instead of stretching each card.
const CARD_GRID = 'grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-3'

const MAX_SHOWN = 200

export function ControlsTab() {
  const { t } = useTranslation()
  const [controls, setControls] = useState<Control[]>([])
  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')
  const [family, setFamily] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [starting, setStarting] = useState(false)
  const [editing, setEditing] = useState<{ control?: Control; creating: boolean } | null>(null)

  useEffect(() => {
    const h = setTimeout(() => setDebounced(search.trim().toLowerCase()), 250)
    return () => clearTimeout(h)
  }, [search])

  const load = useCallback(() => {
    setLoading(true)
    setError(false)
    complianceService
      .listControls()
      .then((c) => setControls(c ?? []))
      .catch(() => setError(true))
      .finally(() => setLoading(false))
  }, [])
  useEffect(() => {
    load()
  }, [load])

  const families = useMemo(() => {
    const s = new Set<string>()
    for (const c of controls) if (c.family) s.add(c.family)
    return [...s].sort()
  }, [controls])

  const filtered = useMemo(() => {
    return controls.filter((c) => {
      if (family && c.family !== family) return false
      if (debounced && !(c.id + ' ' + c.name).toLowerCase().includes(debounced)) return false
      return true
    })
  }, [controls, family, debounced])

  const shown = filtered.slice(0, MAX_SHOWN)

  return (
    <>
      <div className="flex shrink-0 flex-wrap items-center gap-2">
        <div className="relative">
          <Search size={14} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input value={search} onChange={(e) => setSearch(e.target.value)} placeholder={t('compliance.controls.search')} className="w-[260px] pl-8" />
        </div>
        <select value={family} onChange={(e) => setFamily(e.target.value)} className="h-9 rounded-md border border-input bg-background px-2.5 text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring">
          <option value="">{t('compliance.controls.allFamilies')}</option>
          {families.map((f) => (
            <option key={f} value={f}>{f}</option>
          ))}
        </select>
        <Button variant="outline" size="sm" onClick={load} disabled={loading} title={t('compliance.refresh')}>
          <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
        </Button>
        <div className="ml-auto">
          <Button size="sm" onClick={() => setStarting(true)}>
            <Plus size={14} className="mr-1.5" /> {t('compliance.controls.new')}
          </Button>
        </div>
      </div>

      <div className="mt-3 min-h-0 flex-1 overflow-y-auto">
        {loading && controls.length === 0 ? (
          <Center><Loader2 className="h-4 w-4 animate-spin" /> {t('compliance.controls.loading')}</Center>
        ) : error ? (
          <Center><AlertTriangle size={16} className="text-amber-500" /> {t('compliance.controls.loadError')}<Button variant="outline" size="sm" className="ml-2" onClick={load}>{t('compliance.retry')}</Button></Center>
        ) : shown.length === 0 ? (
          <div className="px-6 py-16 text-center text-sm text-muted-foreground">{t('compliance.controls.empty')}</div>
        ) : (
          <div className={CARD_GRID}>
            {shown.map((c) => (
              <div
                key={c.id}
                onClick={() => setEditing({ control: c, creating: false })}
                className={cn(
                  'flex cursor-pointer flex-col rounded-xl border border-border bg-card p-4 transition-colors hover:border-primary/40',
                  c.locked && 'opacity-60',
                )}
              >
                <div className="flex items-start justify-between gap-2">
                  <div className="flex min-w-0 items-center gap-1.5">
                    <ShieldCheck size={13} className="shrink-0 text-muted-foreground" />
                    <span className="truncate font-mono text-[11px] text-muted-foreground">{c.id}</span>
                  </div>
                                    {c.locked && (
                    <span
                      className="inline-flex shrink-0 items-center gap-1 rounded bg-amber-500/15 px-1.5 py-0.5 text-[9px] font-medium text-amber-600 dark:text-amber-400"
                      title={t('compliance.locked.upsell')}
                    >
                      <Lock size={9} /> {t('compliance.locked.badge')}
                    </span>
                  )}
                </div>

                {/* Two lines, then it stops: with 967 controls, cards that grow
                    with their title turn the grid into a ragged wall. */}
                <h3 className="mt-1.5 line-clamp-2 text-[13px] font-medium leading-snug">{c.name}</h3>

                <div className="mt-auto flex items-center justify-between gap-2 pt-3 text-[11px] text-muted-foreground">
                  <span className="truncate">{c.familyName || c.family || '—'}</span>
                  <span className="shrink-0">{c.scope ? t(`compliance.controls.scopeOpt.${c.scope}`) : '—'}</span>
                </div>
              </div>
            ))}
          </div>
        )}
        {filtered.length > MAX_SHOWN && (
          <div className="px-4 py-4 text-center text-[11px] text-muted-foreground">{t('compliance.controls.showingOf', { shown: MAX_SHOWN, total: filtered.length })}</div>
        )}
      </div>

      {starting && (
        <StartFromModal
          title={t('compliance.controls.new')}
          options={controls.map((c) => ({ id: c.id, name: c.name }))}
          onScratch={() => {
            setStarting(false)
            setEditing({ creating: true })
          }}
          onCopy={(id) => {
            const src = controls.find((c) => c.id === id)
            setStarting(false)
            if (src) setEditing({ control: copyOfControl(src), creating: true })
          }}
          onClose={() => setStarting(false)}
        />
      )}

      {editing && (
        <ControlEditor
          control={editing.control}
          creating={editing.creating}
          onClose={() => setEditing(null)}
          onSaved={() => {
            setEditing(null)
            load()
          }}
        />
      )}
    </>
  )
}

function Center({ children }: { children: React.ReactNode }) {
  return <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">{children}</div>
}
