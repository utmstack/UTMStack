import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, Loader2, X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { complianceService } from '../services/compliance-http.service'
import type { CheckResult, Control, ControlRow, Report } from '../types/compliance.types'
import { ControlStatusBadge } from './ControlStatusBadge'

/** Right-side drawer showing full details of a report control row + its library Control. */
export function ControlDetailDrawer({
  frameworkKey,
  row,
  onClose,
  onStatusChanged,
}: {
  frameworkKey: string
  row: ControlRow
  onClose: () => void
  onStatusChanged?: (updated: Report) => void
}) {
  const { t } = useTranslation()
  const [control, setControl] = useState<Control | null>(null)
  const [loading, setLoading] = useState(true)
  const [notFound, setNotFound] = useState(false)

  useEffect(() => {
    setLoading(true)
    setNotFound(false)
    complianceService
      .getControl(row.controlId)
      .then(setControl)
      .catch(() => setNotFound(true))
      .finally(() => setLoading(false))
  }, [row.controlId])

  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <aside
        className="flex w-full max-w-[560px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-start justify-between gap-4 border-b border-border px-6 py-4">
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <span className="font-mono text-[11px] text-muted-foreground">{row.controlId}</span>
              <ControlStatusBadge frameworkKey={frameworkKey} row={row} onChanged={onStatusChanged} />
            </div>
            <h2 className="mt-1 truncate text-sm font-semibold">{control?.name || row.name}</h2>
          </div>
          <button
            onClick={onClose}
            className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <X size={16} />
          </button>
        </header>

        <div className="min-h-0 flex-1 space-y-5 overflow-y-auto p-6">
          <section className="grid grid-cols-2 gap-3">
            <Stat label={t('compliance.coverageLabel', { defaultValue: 'Coverage' })} value={t('compliance.coverage', { n: row.coverage })} />
            <Stat label={t('compliance.activityLabel', { defaultValue: 'Activity' })} value={t('compliance.activity', { n: row.activity })} />
          </section>
          <EvidenceBlock text={row.evidence} />

          {/* What each check actually returned. The one-line evidence names the
              first failure only; a control failing three of five checks reported
              one, which was enough to work with on screen and not enough to hand
              an auditor. */}
          {row.checks && row.checks.length > 0 && (
            <Block label={t('compliance.checkResults', { defaultValue: 'Check results' })}>
              <ul className="space-y-1.5">
                {row.checks.map((c, i) => (
                  <CheckResultLine key={i} result={c} />
                ))}
              </ul>
            </Block>
          )}

          {row.note && (
            <Block label={t('compliance.verdict', { defaultValue: 'Recorded verdict' })}>
              <p className="whitespace-pre-wrap text-[13px] leading-relaxed text-foreground/90">{row.note}</p>
              <p className="mt-1.5 text-[11px] text-muted-foreground">
                {row.editedBy}
                {row.originalStatus && ` · ${t('compliance.engineSaid', { defaultValue: 'engine said' })} ${t(`compliance.status.${row.originalStatus}`)}`}
              </p>
            </Block>
          )}

          {loading ? (
            <div className="flex items-center justify-center gap-2 py-8 text-sm text-muted-foreground">
              <Loader2 className="h-4 w-4 animate-spin" /> {t('compliance.loading')}
            </div>
          ) : notFound || !control ? (
            <div className="flex items-center justify-center gap-2 py-8 text-sm text-muted-foreground">
              <AlertTriangle size={14} className="text-amber-500" />
              {t('compliance.controls.notInLibrary', { defaultValue: 'Control not found in library' })}
            </div>
          ) : (
            <>
              <section className="grid grid-cols-2 gap-3">
                {control.family && <Stat label={t('compliance.controls.family')} value={`${control.family}${control.familyName ? ` — ${control.familyName}` : ''}`} />}
                {control.scope && <Stat label={t('compliance.controls.scope')} value={t(`compliance.controls.scopeOpt.${control.scope}`, { defaultValue: control.scope })} />}
                {control.strategy && <Stat label={t('compliance.controls.strategy')} value={control.strategy} />}
                {control.source && <Stat label={t('compliance.controls.source', { defaultValue: 'Source' })} value={control.source} />}
              </section>

              {control.statement && (
                <Block label={t('compliance.controls.statement')}>
                  <p className="whitespace-pre-wrap text-[13px] leading-relaxed text-foreground/90">{control.statement}</p>
                </Block>
              )}
              {control.remediation && (
                <Block label={t('compliance.controls.remediation')}>
                  <p className="whitespace-pre-wrap text-[13px] leading-relaxed text-foreground/90">{control.remediation}</p>
                </Block>
              )}

              {control.checks && control.checks.length > 0 && (
                <Block label={t('compliance.controls.checks')}>
                  <ul className="space-y-2">
                    {control.checks.map((c, i) => (
                      <li key={i} className="rounded-md border border-border bg-background/40 p-3">
                        <div className="flex items-baseline gap-2">
                          <span className="font-mono text-[10px] text-muted-foreground">{c.key || '—'}</span>
                          <span className="text-[13px] font-medium">{c.name}</span>
                          {c.todo && <span className="ml-auto rounded bg-sky-500/15 px-1.5 py-0.5 text-[10px] font-semibold text-sky-400">TODO</span>}
                        </div>
                        <div className="mt-1 font-mono text-[10px] text-muted-foreground">
                          {c.dataset ?? 'logs'}
                          {c.dataType ? ` · ${c.dataType}` : ''}
                        </div>
                        {c.filters && c.filters.length > 0 && (
                          <ul className="mt-1.5 space-y-0.5">
                            {c.filters.map((f, fi) => (
                              <li key={fi} className="font-mono text-[10px] text-muted-foreground">
                                {f.field} <span className="text-foreground/70">{f.operator}</span>{' '}
                                {Array.isArray(f.value) ? f.value.join(', ') : String(f.value ?? '')}
                              </li>
                            ))}
                          </ul>
                        )}
                        {c.rule && (
                          <div className="mt-1.5 text-[11px] text-muted-foreground">
                            {c.rule}
                            {c.ruleValue != null && ` = ${c.ruleValue}`}
                          </div>
                        )}
                      </li>
                    ))}
                  </ul>
                </Block>
              )}
            </>
          )}
        </div>
      </aside>
    </div>
  )
}

const OUTCOME_TONE: Record<CheckResult['outcome'], string> = {
  PASSED: 'bg-emerald-500/15 text-emerald-500',
  FAILED: 'bg-red-500/15 text-red-500',
  NOT_APPLICABLE: 'bg-slate-500/15 text-slate-400',
  ERROR: 'bg-amber-500/15 text-amber-500',
}

function CheckResultLine({ result }: { result: CheckResult }) {
  const { t } = useTranslation()
  return (
    <li className="rounded-md border border-border bg-background/40 p-2.5">
      <div className="flex items-baseline gap-2">
        <span className={cn('rounded px-1.5 py-0.5 text-[10px] font-semibold', OUTCOME_TONE[result.outcome])}>
          {t(`compliance.outcome.${result.outcome}`, { defaultValue: result.outcome })}
        </span>
        <span className="min-w-0 flex-1 truncate text-[12px]">{result.name}</span>
      </div>
      <div className="mt-1 font-mono text-[10px] text-muted-foreground">
        {/* A NOT_APPLICABLE check failed on nothing — its data type never
            arrived — so the data type is the explanation, not the hit count. */}
        {result.outcome === 'NOT_APPLICABLE'
          ? t('compliance.noDataFor', { dataType: result.dataType || result.dataset || 'logs', defaultValue: `no ${result.dataType} data in the window` })
          : result.outcome === 'ERROR'
            ? result.error
            : `${result.hits} ${t('compliance.hits', { defaultValue: 'hits' })}${result.required != null ? ` · ${result.rule} ${result.required}` : ''}`}
      </div>
    </li>
  )
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md border border-border bg-background/40 px-3 py-2">
      <div className="text-[10px] uppercase tracking-wide text-muted-foreground">{label}</div>
      <div className="mt-0.5 text-[13px]">{value}</div>
    </div>
  )
}

function Block({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <section>
      <div className="mb-2 text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">{label}</div>
      {children}
    </section>
  )
}

// EvidenceBlock spans the full drawer width, clamps evidence to 2 lines by
// default, and offers a "see more" toggle when the content overflows.
function EvidenceBlock({ text }: { text: string }) {
  const { t } = useTranslation()
  const [expanded, setExpanded] = useState(false)
  const [truncated, setTruncated] = useState(false)
  const bodyRef = useRef<HTMLDivElement | null>(null)

  useLayoutEffect(() => {
    const el = bodyRef.current
    if (!el) return
    // Only meaningful while clamped: scrollHeight > clientHeight means content overflows 2 lines.
    setTruncated(el.scrollHeight > el.clientHeight + 1)
  }, [text])

  const value = text || '—'
  return (
    <section className="rounded-md border border-border bg-background/40 px-3 py-2">
      <div className="text-[10px] uppercase tracking-wide text-muted-foreground">
        {t('compliance.evidenceLabel', { defaultValue: 'Evidence' })}
      </div>
      <div
        ref={bodyRef}
        className={cn('mt-0.5 whitespace-pre-wrap break-words text-[13px]', !expanded && 'line-clamp-2')}
      >
        {value}
      </div>
      {text && (truncated || expanded) && (
        <button
          type="button"
          onClick={() => setExpanded((v) => !v)}
          className="mt-1 text-[11px] font-medium text-primary hover:underline"
        >
          {expanded ? t('compliance.seeLess', { defaultValue: 'See less' }) : t('compliance.seeMore', { defaultValue: 'See more' })}
        </button>
      )}
    </section>
  )
}
