import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Check, Copy, Loader2, Search } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { flatten, isSeverityField, SEVERITY_TONE, HIDDEN_FIELDS } from '../lib/event-fields'

interface Props {
  event?: Record<string, unknown>
  loading?: boolean
  /** Round-trip time for the last test call, for display next to the result. */
  elapsedMs?: number
}

type Tab = 'fields' | 'json'

function formatElapsed(ms: number): string {
  return ms < 1000 ? `${Math.round(ms)}ms` : `${(ms / 1000).toFixed(1)}s`
}

/**
 * Purpose-built result view for a playground test — NOT shared with
 * log-explorer. Shows the parsed event as a flattened, searchable field
 * table (default) or raw JSON, with severity color-coding.
 */
export function ParsedEventView({ event, loading, elapsedMs }: Props) {
  const { t } = useTranslation()
  const [tab, setTab] = useState<Tab>('fields')
  const [copied, setCopied] = useState(false)
  const [search, setSearch] = useState('')

  const rows = useMemo(() => {
    if (!event) return []
    const flat = flatten(event)
    return Object.entries(flat)
      .filter(([k]) => !HIDDEN_FIELDS.has(k))
      .map(([k, v]) => ({ key: k, value: v }))
      .sort((a, b) => a.key.localeCompare(b.key))
  }, [event])

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase()
    if (!q) return rows
    return rows.filter((r) => r.key.toLowerCase().includes(q) || String(r.value).toLowerCase().includes(q))
  }, [rows, search])

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center gap-2 rounded-md border border-border bg-card px-3 py-10 text-sm text-muted-foreground">
        <Loader2 size={14} className="animate-spin" /> {t('playground.result.loading')}
      </div>
    )
  }

  if (!event) {
    return (
      <div className="flex h-full items-center justify-center rounded-md border border-dashed border-border px-3 py-10 text-center text-sm text-muted-foreground">
        {t('playground.result.empty')}
      </div>
    )
  }

  const json = JSON.stringify(event, null, 2)
  const copy = () => {
    void navigator.clipboard.writeText(json)
    setCopied(true)
    toast.success(t('playground.result.copied'))
    setTimeout(() => setCopied(false), 1500)
  }

  return (
    <div className="flex h-full flex-col overflow-hidden rounded-md border border-border bg-card">
      <div className="flex items-center justify-between gap-3 border-b border-border/60 px-3 py-2">
        <div className="flex items-center gap-4">
          <TabBtn id="fields" current={tab} onChange={setTab}>
            {t('playground.result.tabs.fields')}
          </TabBtn>
          <TabBtn id="json" current={tab} onChange={setTab}>
            {t('playground.result.tabs.json')}
          </TabBtn>
        </div>
        <div className="flex items-center gap-3">
          {elapsedMs != null && (
            <span className="text-[11px] text-muted-foreground" title={t('playground.result.elapsedHint')}>
              {t('playground.result.elapsed', { time: formatElapsed(elapsedMs) })}
            </span>
          )}
          <span className="text-[11px] text-muted-foreground">{t('playground.result.fieldsCount', { count: rows.length })}</span>
          <button
            type="button"
            onClick={copy}
            className="flex h-6 items-center gap-1.5 rounded-md px-1.5 text-[11px] text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            {copied ? <Check size={11} className="text-emerald-500" /> : <Copy size={11} />}
            {copied ? t('playground.result.copied') : t('playground.result.copy')}
          </button>
        </div>
      </div>

      {tab === 'fields' && (
        <div className="border-b border-border/60 px-3 py-2">
          <div className="relative">
            <Search size={12} className="absolute left-2 top-1/2 -translate-y-1/2 text-muted-foreground" />
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder={t('playground.result.searchPlaceholder')}
              className="h-7 w-full rounded-md border border-input bg-background/40 pl-6 pr-2 text-[11px] focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            />
          </div>
        </div>
      )}

      <div className="min-h-0 flex-1 overflow-auto p-3">
        {tab === 'fields' ? (
          filtered.length === 0 ? (
            <p className="px-1 py-6 text-center text-[12px] text-muted-foreground">{t('playground.result.noMatch')}</p>
          ) : (
            <div className="overflow-hidden rounded-md border border-border">
              {filtered.map((row, i) => {
                const sevTone = isSeverityField(row.key) ? SEVERITY_TONE[String(row.value).toLowerCase()] : undefined
                return (
                  <div
                    key={row.key}
                    className={cn(
                      'grid grid-cols-[minmax(140px,220px)_1fr] items-start gap-3 px-3 py-1.5 text-[12px] leading-relaxed',
                      i < filtered.length - 1 && 'border-b border-border/60'
                    )}
                  >
                    <div className="truncate font-mono text-muted-foreground" title={row.key}>
                      {row.key}
                    </div>
                    {sevTone ? (
                      <div>
                        <span className={cn('inline-flex rounded px-1.5 py-0.5 font-mono text-[11px]', sevTone)}>
                          {String(row.value)}
                        </span>
                      </div>
                    ) : (
                      <div className="break-all font-mono">{String(row.value)}</div>
                    )}
                  </div>
                )
              })}
            </div>
          )
        ) : (
          <pre className="overflow-x-auto font-mono text-[11px] leading-relaxed">{json}</pre>
        )}
      </div>
    </div>
  )
}

function TabBtn({
  id,
  current,
  onChange,
  children,
}: {
  id: Tab
  current: Tab
  onChange: (t: Tab) => void
  children: React.ReactNode
}) {
  const active = id === current
  return (
    <button
      type="button"
      onClick={() => onChange(id)}
      className={cn(
        'relative -mb-2 border-b-2 pb-2 text-xs transition-colors',
        active ? 'border-foreground font-medium text-foreground' : 'border-transparent text-muted-foreground hover:text-foreground'
      )}
    >
      {children}
    </button>
  )
}
