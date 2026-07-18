import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ChevronDown, ChevronRight, Loader2 } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { flatten, isSeverityField, SEVERITY_TONE, HIDDEN_FIELDS } from '../lib/event-fields'

interface Props {
  alerts?: unknown[]
  loading?: boolean
}

// Common keys the underlying rule engine may use for an alert's display name,
// checked in order against the flattened record (dotted paths included).
const NAME_KEYS = ['name', 'ruleName', 'rule.name', 'alertName']

function findFirst(flat: Record<string, unknown>, keys: string[]): unknown {
  for (const k of keys) if (flat[k] != null) return flat[k]
  return undefined
}

/** First flattened key/value pair that looks like a severity/level field. */
function findSeverity(flat: Record<string, unknown>): { key: string; value: unknown } | undefined {
  const key = Object.keys(flat).find((k) => isSeverityField(k))
  return key ? { key, value: flat[key] } : undefined
}

/**
 * Renders the `alerts` half of a rule-mode playground test — a collapsible
 * row per alert (name + severity chip in the header, chevron to expand) whose
 * body is a flattened field table sharing the exact styling and helpers used
 * by `ParsedEventView`'s field table (no icons, no extra tinting on field
 * names — only the severity value itself gets a colored pill).
 */
export function AlertsListView({ alerts, loading }: Props) {
  const { t } = useTranslation()

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center gap-2 rounded-md border border-border bg-card px-3 py-10 text-sm text-muted-foreground">
        <Loader2 size={14} className="animate-spin" />
      </div>
    )
  }

  if (!alerts) {
    return (
      <div className="flex h-full items-center justify-center rounded-md border border-dashed border-border px-3 py-10 text-center text-sm text-muted-foreground">
        {t('playground.alerts.empty')}
      </div>
    )
  }

  if (alerts.length === 0) {
    return (
      <div className="flex h-full items-center justify-center rounded-md border border-border bg-card px-3 py-10 text-center text-sm text-muted-foreground">
        {t('playground.alerts.none')}
      </div>
    )
  }

  return (
    <div className="flex h-full flex-col overflow-hidden rounded-md border border-border bg-card">
      <div className="flex items-center justify-between gap-3 border-b border-border/60 px-3 py-2">
        <span className="text-[11px] text-muted-foreground">{t('playground.alerts.count', { count: alerts.length })}</span>
      </div>
      <div className="min-h-0 flex-1 overflow-auto p-3">
        <div className="space-y-2">
          {alerts.map((alert, i) => (
            <AlertRow key={i} alert={alert} />
          ))}
        </div>
      </div>
    </div>
  )
}

function AlertRow({ alert }: { alert: unknown }) {
  const [open, setOpen] = useState(false)

  const flat = useMemo(() => flatten(alert), [alert])
  const name = useMemo(() => {
    const v = findFirst(flat, NAME_KEYS)
    return v != null ? String(v) : undefined
  }, [flat])
  const severity = useMemo(() => findSeverity(flat), [flat])
  const sevTone = severity ? SEVERITY_TONE[String(severity.value).toLowerCase()] : undefined

  const rows = useMemo(
    () =>
      Object.entries(flat)
        .filter(([k]) => !HIDDEN_FIELDS.has(k))
        .map(([k, v]) => ({ key: k, value: v }))
        .sort((a, b) => a.key.localeCompare(b.key)),
    [flat],
  )

  return (
    <div className="overflow-hidden rounded-md border border-border">
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        className="flex w-full items-center gap-2 px-3 py-2 text-left text-xs hover:bg-muted/40"
      >
        {open ? <ChevronDown size={13} className="shrink-0 text-muted-foreground" /> : <ChevronRight size={13} className="shrink-0 text-muted-foreground" />}
        <span className="min-w-0 flex-1 truncate font-medium">{name ?? '—'}</span>
        {severity &&
          (sevTone ? (
            <span className={cn('inline-flex shrink-0 rounded px-1.5 py-0.5 font-mono text-[11px]', sevTone)}>{String(severity.value)}</span>
          ) : (
            <span className="shrink-0 font-mono text-[11px] text-muted-foreground">{String(severity.value)}</span>
          ))}
      </button>
      {open && (
        <div className="overflow-hidden border-t border-border">
          {rows.map((row, i) => {
            const rowSevTone = isSeverityField(row.key) ? SEVERITY_TONE[String(row.value).toLowerCase()] : undefined
            return (
              <div
                key={row.key}
                className={cn(
                  'grid grid-cols-[minmax(140px,220px)_1fr] items-start gap-3 px-3 py-1.5 text-[12px] leading-relaxed',
                  i < rows.length - 1 && 'border-b border-border/60',
                )}
              >
                <div className="truncate font-mono text-muted-foreground" title={row.key}>
                  {row.key}
                </div>
                {rowSevTone ? (
                  <div>
                    <span className={cn('inline-flex rounded px-1.5 py-0.5 font-mono text-[11px]', rowSevTone)}>{String(row.value)}</span>
                  </div>
                ) : (
                  <div className="break-all font-mono">{String(row.value)}</div>
                )}
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
