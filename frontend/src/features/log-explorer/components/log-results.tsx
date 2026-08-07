import { memo, useCallback, useMemo, useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { ChevronRight, Copy, Crosshair, Minus, Plus, X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import type { FilterType, LogDocument } from '../types/log-explorer.types'
import { MSG_FIELDS, SRC_FIELDS, docPreview, flattenDoc, pick } from '../domain/flatten'

/**
 * Shared "Discover-style" log results rendering: the table row + the expandable
 * document detail (parsed fields / JSON). Used by the Log Explorer and by the
 * alert drawer's Related events tab so logs look identical everywhere. Pass
 * `onAdd` to enable the per-field filter +/- buttons (omit for a read-only view).
 */

const TS = '@timestamp'

const LEVEL_FIELDS = ['log.level', 'severity', 'level', 'event.severity', 'logx.severity']

const LEVEL_TONE: Record<string, { dot: string; tone: string }> = {
  critical: { dot: 'bg-red-500', tone: 'text-red-500' },
  high: { dot: 'bg-red-500', tone: 'text-red-500' },
  error: { dot: 'bg-orange-500', tone: 'text-orange-500' },
  warn: { dot: 'bg-amber-500', tone: 'text-amber-500' },
  warning: { dot: 'bg-amber-500', tone: 'text-amber-500' },
  medium: { dot: 'bg-amber-500', tone: 'text-amber-500' },
  info: { dot: 'bg-sky-500', tone: 'text-sky-500' },
  low: { dot: 'bg-sky-500', tone: 'text-sky-500' },
  debug: { dot: 'bg-muted-foreground', tone: 'text-muted-foreground' },
}

function shortTime(iso: string) {
  const d = new Date(iso)
  return Number.isNaN(d.getTime())
    ? iso
    : d.toLocaleString(undefined, { month: 'short', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function absTimestamp(iso: string) {
  const d = new Date(iso)
  return Number.isNaN(d.getTime())
    ? iso
    : d.toLocaleString(undefined, { month: 'short', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

// Grid columns. Manual mode (user picked columns): time + each picked column (last
// flexes). Default mode: time + source + auto-detected important columns + a
// flexible message column.
function gridTemplate(columns: string[], autoColumns: string[] = []): string {
  if (columns.length > 0) return `20px 3px 168px ${columns.map(() => 'minmax(120px, 1fr)').join(' ')}`
  const auto = autoColumns.map(() => 'minmax(96px, 0.7fr)').join(' ')
  return `20px 3px 168px 120px ${auto ? auto + ' ' : ''}minmax(0, 1fr)`
}

function colValue(flat: Record<string, unknown>, c: string): string {
  const v = flat[c]
  if (v == null || v === '') return '—'
  if (c === TS) return shortTime(String(v))
  return String(v)
}

function ResultsHeaderImpl({
  columns,
  autoColumns = [],
  onRemoveColumn,
}: {
  columns: string[]
  autoColumns?: string[]
  onRemoveColumn?: (c: string) => void
}) {
  const { t } = useTranslation()
  return (
    <div
      className="sticky top-0 z-10 grid items-center gap-3 border-b border-border/70 bg-card px-4 py-1.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground"
      style={{ gridTemplateColumns: gridTemplate(columns, autoColumns) }}
    >
      <div />
      <div />
      <div>{t('logExplorer.results.time')}</div>
      {columns.length === 0 ? (
        <>
          <div>{t('logExplorer.results.source')}</div>
          {autoColumns.map((c,i) => (
            <div key={i} className="truncate" title={c}>
              {fieldLabel(c)}
            </div>
          ))}
          <div>{t('logExplorer.results.message')}</div>
        </>
      ) : (
        columns.map((c,i) => (
          <div key={i} className="group flex min-w-0 items-center gap-1">
            <span className="truncate" title={c}>
              {fieldLabel(c)}
            </span>
            {onRemoveColumn && (
              <button
                onClick={() => onRemoveColumn(c)}
                title={t('logExplorer.results.removeColumn', { field: c })}
                className="shrink-0 opacity-0 transition-opacity hover:text-foreground group-hover:opacity-100"
              >
                <X size={11} />
              </button>
            )}
          </div>
        ))
      )}
    </div>
  )
}

export const ResultsHeader = memo(ResultsHeaderImpl)

// Short, readable column header from a field path: "origin.ip" → "origin ip".
function fieldLabel(field: string): string {
  return field.replace(/\./g, ' ')
}

function ResultRowImpl({
  index,
  doc,
  columns,
  autoColumns = [],
  expanded,
  onToggle,
  onAdd,
  onSurrounding,
}: {
  index: number
  doc: LogDocument
  columns: string[]
  autoColumns?: string[]
  expanded: boolean
  onToggle: (index: number) => void
  onAdd?: (f: FilterType) => void
  onSurrounding?: (ts: string, srcField?: string, srcVal?: string) => void
}) {
  const flat = useMemo(() => flattenDoc(doc), [doc])
  const ts = (flat[TS] as string) ?? ''
  const source = pick(flat, SRC_FIELDS) ?? '—'
  const level = (pick(flat, LEVEL_FIELDS) ?? '').toLowerCase()
  const tone = LEVEL_TONE[level] ?? { dot: 'bg-muted-foreground/50', tone: 'text-muted-foreground' }
  const message = pick(flat, MSG_FIELDS)
  const preview = useMemo(() => (columns.length > 0 || message ? null : docPreview(flat)), [columns.length, message, flat])

  return (
    <>
      <div
        onClick={() => onToggle(index)}
        className={cn(
          'grid cursor-pointer items-center gap-3 border-b border-border/40 px-4 py-1 text-xs leading-tight transition-colors last:border-b-0',
          expanded ? 'bg-muted/30' : 'hover:bg-muted/20'
        )}
        style={{ gridTemplateColumns: gridTemplate(columns, autoColumns) }}
      >
        <ChevronRight size={13} className={cn('text-muted-foreground/60 transition-transform', expanded && 'rotate-90 text-foreground')} />
        <span className={cn('h-3.5 w-[3px] rounded-full', tone.dot)} />
        <div className="font-mono tabular-nums text-muted-foreground">{ts ? shortTime(ts) : '—'}</div>
        {columns.length > 0 ? (
          columns.map((c,i) => (
            <div key={i} className="truncate font-mono text-foreground/85" title={colValue(flat, c)}>
              {colValue(flat, c)}
            </div>
          ))
        ) : (
          <>
            <div className="truncate font-mono text-foreground/70">{source}</div>
            {autoColumns.map((c,i) => {
              const val = colValue(flat, c)
              return (
                <div key={i} className={cn('truncate font-mono', val === '—' ? 'text-muted-foreground/40' : 'text-foreground/85')} title={val}>
                  {val}
                </div>
              )
            })}
            {message ? (
              <div className="truncate text-foreground">{message}</div>
            ) : (
              <div className="flex items-center overflow-hidden whitespace-nowrap">
                {preview!.map(([k, v], idx) => (
                  <span key={idx} className="flex shrink-0 items-center">
                    {idx > 0 && <span className="px-2.5 text-border">·</span>}
                    <span className="text-muted-foreground">{k}</span>
                    <span className="ml-1.5 font-mono text-foreground">{v}</span>
                  </span>
                ))}
              </div>
            )}
          </>
        )}
      </div>
      {expanded && <ExpandedPanel flat={flat} doc={doc} onAdd={onAdd} onSurrounding={onSurrounding} />}
    </>
  )
}

type DetailTab = 'fields' | 'json'

function ExpandedPanel({
  flat,
  doc,
  onAdd,
  onSurrounding,
}: {
  flat: Record<string, unknown>
  doc: LogDocument
  onAdd?: (f: FilterType) => void
  onSurrounding?: (ts: string, srcField?: string, srcVal?: string) => void
}) {
  const { t } = useTranslation()
  const [tab, setTab] = useState<DetailTab>('fields')
  const ts = (flat[TS] as string) ?? ''
  const srcField = SRC_FIELDS.find((f) => flat[f] != null)
  const srcVal = srcField != null ? String(flat[srcField]) : undefined
  const entries = Object.entries(flat).sort(([a], [b]) => a.localeCompare(b))

  return (
    <div className="border-b border-l-2 border-border/50 border-l-sky-500/50 bg-muted/15 last:border-b-0">
      <div className="flex items-center justify-between gap-4 border-b border-border/40 px-5 py-2.5">
        <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
          {ts && <span className="font-mono">{absTimestamp(ts)}</span>}
          <span className="font-mono">{t('logExplorer.detail.fieldsCount', { count: entries.length })}</span>
        </div>
        <div className="flex items-center gap-1.5">
          {onSurrounding && ts && (
            <button
              onClick={() => onSurrounding(ts, srcField, srcVal)}
              title={t('logExplorer.detail.surroundingHint')}
              className="flex h-7 items-center gap-1.5 rounded-md px-2 text-[11px] text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <Crosshair size={12} /> {t('logExplorer.detail.surrounding')}
            </button>
          )}
          <button
            onClick={() => {
              void navigator.clipboard.writeText(JSON.stringify(doc, null, 2))
              toast.success(t('logExplorer.detail.copied'))
            }}
            className="flex h-7 items-center gap-1.5 rounded-md px-2 text-[11px] text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <Copy size={12} /> {t('logExplorer.detail.copyJson')}
          </button>
        </div>
      </div>

      <div className="flex items-center gap-4 border-b border-border/60 px-5">
        <DetailTabBtn id="fields" current={tab} onChange={setTab}>
          {t('logExplorer.detail.parsedFields')}
        </DetailTabBtn>
        <DetailTabBtn id="json" current={tab} onChange={setTab}>
          {t('logExplorer.detail.json')}
        </DetailTabBtn>
      </div>

      <div className="p-5">
        {tab === 'fields' ? (
          <div className="overflow-hidden rounded-md border border-border bg-card">
            {entries.map(([k, v], i) => (
              <div
                key={`${k}-${i}`}
                className={cn(
                  'group grid items-center gap-4 px-4 py-2 text-[13px] leading-relaxed hover:bg-muted/30',
                  onAdd ? 'grid-cols-[260px_1fr_60px]' : 'grid-cols-[260px_1fr]',
                  i < entries.length - 1 && 'border-b border-border/60'
                )}
              >
                <div className="truncate font-mono text-xs text-muted-foreground">{k}</div>
                <div className="break-all font-mono text-xs">{String(v)}</div>
                {onAdd && (
                  <div className="flex justify-end gap-0.5 opacity-0 transition-opacity group-hover:opacity-100">
                    <button
                      title={t('logExplorer.detail.filterFor')}
                      onClick={() => onAdd({ field: k, operator: 'IS', value: String(v) })}
                      className="flex h-5 w-5 items-center justify-center rounded text-emerald-500 hover:bg-emerald-500/15"
                    >
                      <Plus size={10} />
                    </button>
                    <button
                      title={t('logExplorer.detail.filterOut')}
                      onClick={() => onAdd({ field: k, operator: 'IS_NOT', value: String(v) })}
                      className="flex h-5 w-5 items-center justify-center rounded text-red-500 hover:bg-red-500/15"
                    >
                      <Minus size={10} />
                    </button>
                  </div>
                )}
              </div>
            ))}
          </div>
        ) : (
          <pre className="overflow-x-auto rounded-md border border-border bg-card p-3 font-mono text-[11px] leading-relaxed">
            {JSON.stringify(doc, null, 2)}
          </pre>
        )}
      </div>
    </div>
  )
}

function DetailTabBtn({
  id,
  current,
  onChange,
  children,
}: {
  id: DetailTab
  current: DetailTab
  onChange: (t: DetailTab) => void
  children: ReactNode
}) {
  const active = id === current
  return (
    <button
      onClick={() => onChange(id)}
      className={cn(
        'relative -mb-px border-b-2 py-2 text-xs transition-colors',
        active ? 'border-foreground text-foreground' : 'border-transparent text-muted-foreground hover:text-foreground'
      )}
    >
      {children}
    </button>
  )
}

export const ResultRow = memo(ResultRowImpl)

/**
 * Self-contained list: header + expandable rows with their own expand state.
 * Read-only by default (no filter buttons, no column removal).
 */
export function LogResults({
  docs,
  columns = [],
  onAdd,
  emptyText,
}: {
  docs: LogDocument[]
  columns?: string[]
  onAdd?: (f: FilterType) => void
  emptyText?: string
}) {
  const [expanded, setExpanded] = useState<number | null>(null)
  const toggle = useCallback((i: number) => setExpanded((prev) => (prev === i ? null : i)), [])
  return (
    <div className="overflow-auto rounded-lg border border-border bg-card">
      {docs.length === 0 ? (
        <div className="px-6 py-12 text-center text-sm text-muted-foreground">{emptyText ?? '—'}</div>
      ) : (
        docs.map((doc, i) => (
          <ResultRow
            key={i}
            index={i}
            doc={doc}
            columns={columns}
            expanded={expanded === i}
            onToggle={toggle}
            onAdd={onAdd}
          />
        ))
      )}
    </div>
  )
}
