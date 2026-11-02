import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Check, Loader2, Play, Search, X } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Input } from '@/shared/components/ui/input'
import { Textarea } from '@/shared/components/ui/textarea'
import { Button } from '@/shared/components/ui/button'
import { alertsHttpService } from '@/features/alerts/services/alerts-http.service'
import { datasourcesHttpService } from '@/features/datasources/services/datasources-http.service'
import { relativeTime, sevKey, SEV_BADGE } from '@/features/alerts/lib/alert-meta'
import type { Alert } from '@/features/alerts/types/alert.types'
import { soarFlowsService } from '../services/soar-flows.service'
import type { Flow } from '../types/soar.types'

const AGENT_SAME = '__same__'

type Source = 'list' | 'raw'

/**
 * Manual flow trigger: pick the alert that "raised" the flow — either a real
 * stored alert (embedded searchable table) or raw pasted alert JSON — and the
 * agent to execute on (default: the alert's own agent). Flow conditions are
 * ignored; a manual trigger runs the flow now.
 */
export function FlowTriggerModal({ flow, onClose }: { flow: Flow; onClose: () => void }) {
  const { t } = useTranslation()
  const [source, setSource] = useState<Source>('list')
  const [query, setQuery] = useState('')
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [total, setTotal] = useState(0)
  const [listLoading, setListLoading] = useState(true)
  const [picked, setPicked] = useState<Alert | null>(null)
  const [raw, setRaw] = useState('')
  const [agents, setAgents] = useState<string[]>([])
  const [agent, setAgent] = useState(AGENT_SAME)
  const [running, setRunning] = useState(false)
  const PAGE = 15

  useEffect(() => {
    datasourcesHttpService
      .list({ page: 1, size: 1000, kind: 'agent' })
      .then((r) => setAgents((r.items ?? []).map((d) => d.name)))
      .catch(() => setAgents([]))
  }, [])

  const load = useCallback(() => {
      setListLoading(true)
      alertsHttpService
        .list({ page: 0, size: PAGE, filters: query.trim() ? [{ field: 'name', operator: 'CONTAIN', value: query.trim() }] : [] })
        .then((r) => {
          setAlerts(r.data ?? [])
          setTotal(r.total ?? 0)
        })
        .catch(() => {
          setAlerts([])
          setTotal(0)
        })
        .finally(() => setListLoading(false))
  }, [query])

  useEffect(() => {
    const h = setTimeout(() => load(), 250)
    return () => clearTimeout(h)
  }, [load])

  const doTrigger = async () => {
    let payload: unknown
    if (source === 'list') {
      if (!picked) return
      payload = picked
    } else {
      try {
        payload = JSON.parse(raw)
      } catch {
        toast.error(t('soar.trigger.invalidJson'))
        return
      }
    }
    setRunning(true)
    try {
      await soarFlowsService.trigger(flow.relPath, payload as Alert, agent === AGENT_SAME ? '' : agent)
      toast.success(t('soar.trigger.ok', { name: flow.name }))
      onClose()
    } catch {
      toast.error(t('soar.trigger.fail'))
    } finally {
      setRunning(false)
    }
  }

  const canRun = source === 'list' ? !!picked : raw.trim() !== ''

  return (
    <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm" onClick={onClose}>
      <div className="flex max-h-[80vh] w-full max-w-lg flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl" onClick={(e) => e.stopPropagation()}>
        <header className="flex shrink-0 items-center gap-3 border-b border-border px-5 py-4">
          <h2 className="min-w-0 flex-1 truncate text-base font-semibold">{t('soar.trigger.title', { name: flow.name })}</h2>
          <button onClick={onClose} aria-label={t('common.cancel')} className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground">
            <X size={16} />
          </button>
        </header>

        <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto p-5">
          {/* alert source */}
          <div>
            <div className="mb-2 inline-flex rounded-md border border-border p-0.5">
              {(['list', 'raw'] as Source[]).map((s) => (
                <button key={s} onClick={() => setSource(s)} className={cn('rounded px-2.5 py-1 text-xs transition-colors', source === s ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground')}>
                  {s === 'list' ? t('soar.trigger.fromList') : t('soar.trigger.rawJson')}
                </button>
              ))}
            </div>

            {source === 'list' ? (
              <>
                <div className="relative mb-2">
                  <Search size={14} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
                  <Input value={query} onChange={(e) => setQuery(e.target.value)} placeholder={t('soar.trigger.searchAlerts')} className="pl-8" />
                </div>
                <div className="max-h-60 overflow-y-auto rounded-lg border border-border">
                  {listLoading ? (
                    <div className="flex items-center justify-center gap-2 py-8 text-[12px] text-muted-foreground">
                      <Loader2 size={14} className="animate-spin" /> {t('soar.loading')}
                    </div>
                  ) : alerts.length === 0 ? (
                    <div className="py-8 text-center text-[12px] text-muted-foreground">{t('soar.trigger.noAlerts')}</div>
                  ) : (
                    alerts.map((a) => {
                      const on = picked?.id === a.id
                      return (
                        <button key={a.id} onClick={() => setPicked(on ? null : a)} className={cn('flex w-full items-center gap-2 border-b border-border px-3 py-2 text-left last:border-b-0 hover:bg-muted/50', on && 'bg-primary/10 hover:bg-primary/10')}>
                          <span className={cn('shrink-0 rounded px-1.5 py-0.5 text-[10px] font-semibold uppercase', SEV_BADGE[sevKey(a)])}>{a.severity ?? ''}</span>
                          <span className="min-w-0 flex-1 truncate text-[12px]">{a.name}</span>
                          <span className="shrink-0 font-mono text-[10px] text-muted-foreground">{relativeTime(a['@timestamp'])}</span>
                          <Check size={14} className={cn('shrink-0 text-primary', on ? 'opacity-100' : 'opacity-0')} />
                        </button>
                      )
                    })
                  )}
                </div>
                {total > alerts.length && <div className="mt-1 text-[11px] text-muted-foreground">{t('soar.trigger.showing', { shown: alerts.length, total })}</div>}
              </>
            ) : (
              <Textarea value={raw} onChange={(e) => setRaw(e.target.value)} rows={6} maxRows={12} placeholder='{ "name": "…", "target": { "host": "…" } }' className="font-mono text-[11px]" />
            )}
          </div>

          {/* agent */}
          <div>
            <div className="mb-1.5 text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">{t('soar.trigger.agent')}</div>
            <select value={agent} onChange={(e) => setAgent(e.target.value)} className="h-8 w-full rounded-md border border-input bg-background px-2 text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring">
              <option value={AGENT_SAME}>{t('soar.trigger.sameAsAlert')}</option>
              {agents.map((a) => (
                <option key={a} value={a}>{a}</option>
              ))}
            </select>
          </div>
        </div>

        <footer className="flex shrink-0 items-center justify-end gap-2 border-t border-border px-5 py-3">
          <Button variant="outline" size="sm" onClick={onClose}>{t('common.cancel')}</Button>
          <Button size="sm" onClick={() => void doTrigger()} disabled={!canRun || running}>
            {running ? <Loader2 size={14} className="mr-1.5 animate-spin" /> : <Play size={14} className="mr-1.5" />}
            {t('soar.trigger.run')}
          </Button>
        </footer>
      </div>
    </div>
  )
}
