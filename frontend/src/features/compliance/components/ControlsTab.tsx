import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, Loader2, Lock, Plus, RefreshCw, Search, ShieldCheck } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { complianceService } from '../services/compliance-http.service'
import type { Control } from '../types/compliance.types'
import { ControlEditor } from './ControlEditor'

const COLS = '120px minmax(180px,1fr) 130px 110px 70px 56px'
const MAX_SHOWN = 200

export function ControlsTab() {
  const { t } = useTranslation()
  const [controls, setControls] = useState<Control[]>([])
  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')
  const [family, setFamily] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
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

  const toggle = async (c: Control) => {
    if (c.locked) {
      toast.error(t('compliance.locked.upsell'))
      return
    }
    const next = !c.enabled
    setControls((list) => list.map((x) => (x.id === c.id ? { ...x, enabled: next } : x)))
    try {
      await complianceService.setControlEnabled(c.id, next)
    } catch {
      setControls((list) => list.map((x) => (x.id === c.id ? { ...x, enabled: c.enabled } : x)))
      toast.error(t('compliance.controls.toast.toggleError'))
    }
  }

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
          <Button size="sm" onClick={() => setEditing({ creating: true })}>
            <Plus size={14} className="mr-1.5" /> {t('compliance.controls.new')}
          </Button>
        </div>
      </div>

      <div className="mt-3 flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-border bg-card">
        <div className="grid items-center gap-3 border-b border-border bg-muted/30 px-4 py-2.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground" style={{ gridTemplateColumns: COLS }}>
          <div>{t('compliance.controls.id')}</div>
          <div>{t('compliance.controls.name')}</div>
          <div>{t('compliance.controls.family')}</div>
          <div>{t('compliance.controls.scope')}</div>
          <div className="text-center">{t('compliance.enabled')}</div>
          <div />
        </div>
        <div className="min-h-0 flex-1 overflow-y-auto">
          {loading && controls.length === 0 ? (
            <Center><Loader2 className="h-4 w-4 animate-spin" /> {t('compliance.controls.loading')}</Center>
          ) : error ? (
            <Center><AlertTriangle size={16} className="text-amber-500" /> {t('compliance.controls.loadError')}<Button variant="outline" size="sm" className="ml-2" onClick={load}>{t('compliance.retry')}</Button></Center>
          ) : shown.length === 0 ? (
            <div className="px-6 py-16 text-center text-sm text-muted-foreground">{t('compliance.controls.empty')}</div>
          ) : (
            shown.map((c) => (
              <div key={c.id} className={cn('grid cursor-pointer items-center gap-3 border-b border-border px-4 py-2.5 text-sm transition-colors last:border-0 hover:bg-muted/40', c.locked && 'opacity-60')} style={{ gridTemplateColumns: COLS }} onClick={() => setEditing({ control: c, creating: false })}>
                <div className="flex items-center gap-1.5">
                  <ShieldCheck size={13} className="shrink-0 text-muted-foreground" />
                  <span className="truncate font-mono text-[12px]">{c.id}</span>
                </div>
                <span className="truncate text-[13px]">{c.name}</span>
                <span className="truncate text-[11px] text-muted-foreground">{c.familyName || c.family || '—'}</span>
                <span className="text-[11px] text-muted-foreground">{c.scope ? t(`compliance.controls.scopeOpt.${c.scope}`) : '—'}</span>
                <div className="flex justify-center" onClick={(e) => e.stopPropagation()}>
                  {c.locked ? (
                    <span className="inline-flex items-center gap-1 rounded bg-amber-500/15 px-1.5 py-0.5 text-[9px] font-medium text-amber-600 dark:text-amber-400" title={t('compliance.locked.upsell')}>
                      <Lock size={9} /> {t('compliance.locked.badge')}
                    </span>
                  ) : (
                    <Toggle checked={c.enabled} onChange={() => toggle(c)} />
                  )}
                </div>
                <div className="flex justify-end">{c.system && !c.locked && <Lock size={11} className="text-muted-foreground/60" />}</div>
              </div>
            ))
          )}
          {filtered.length > MAX_SHOWN && (
            <div className="px-4 py-3 text-center text-[11px] text-muted-foreground">{t('compliance.controls.showingOf', { shown: MAX_SHOWN, total: filtered.length })}</div>
          )}
        </div>
      </div>

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

function Toggle({ checked, onChange }: { checked: boolean; onChange: () => void }) {
  return (
    <button type="button" role="switch" aria-checked={checked} onClick={onChange} className={cn('relative inline-flex h-5 w-9 shrink-0 items-center rounded-full transition-colors', checked ? 'bg-primary' : 'bg-muted-foreground/30')}>
      <span className={cn('inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform', checked ? 'translate-x-4' : 'translate-x-0.5')} />
    </button>
  )
}

function Center({ children }: { children: React.ReactNode }) {
  return <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">{children}</div>
}
