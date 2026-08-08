import { useTranslation } from 'react-i18next'
import { AlertTriangle } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { useDateFormat } from '@/shared/lib/datetime'
import {
  isEditStale,
  isOverridden,
  type ComplianceStatus,
  type ControlRow,
  type Report,
  type ReportSummary,
} from '../types/compliance.types'
import { ControlStatusBadge } from './ControlStatusBadge'
import { ControlNoteButton } from './ControlNoteButton'

export function scoreTone(score: number): string {
  if (score >= 80) return 'text-emerald-500'
  if (score >= 50) return 'text-amber-500'
  return 'text-red-500'
}

export function scoreBg(score: number): string {
  if (score >= 80) return 'bg-emerald-500'
  if (score >= 50) return 'bg-amber-500'
  return 'bg-red-500'
}

export const STATUS_TONE: Record<ComplianceStatus, string> = {
  COMPLIANT: 'bg-emerald-500/15 text-emerald-500',
  NON_COMPLIANT: 'bg-red-500/15 text-red-500',
  AT_RISK: 'bg-amber-500/15 text-amber-500',
  NOT_COVERED: 'bg-zinc-500/15 text-zinc-400',
  NOT_EVALUATED: 'bg-slate-500/15 text-slate-400',
  PENDING: 'bg-sky-500/15 text-sky-400',
  OUT_OF_SCOPE: 'bg-violet-500/15 text-violet-400',
}

const EMPTY_SUMMARY: ReportSummary = {
  compliantPct: 0,
  total: 0,
  evaluated: 0,
  compliant: 0,
  nonCompliant: 0,
  atRisk: 0,
  notCovered: 0,
  notEvaluated: 0,
  pending: 0,
  outOfScope: 0,
}

/**
 * Renders the report: score, then the framework's own shape — sections holding
 * requirements, each resolving to the controls that answer it.
 *
 * Controls arrive flat and are looked up by id, because the crosswalk points
 * many requirements at the same control.
 */
export function ReportView({
  report,
  onControlClick,
  onChanged,
}: {
  report: Report
  onControlClick?: (row: ControlRow) => void
  /** Called after a row-level edit so the parent can take the returned report. */
  onChanged?: (updated: Report) => void
}) {
  const { t } = useTranslation()
  const df = useDateFormat()
  const s = report.summary ?? EMPTY_SUMMARY
  const byId = new Map((report.controls ?? []).map((c) => [c.controlId, c]))

  return (
    <div className="space-y-5">
      <div className="rounded-xl border border-border bg-background/40 p-5">
        <div className="flex items-center gap-5">
          <div className={cn('text-4xl font-bold tabular-nums', scoreTone(s.compliantPct))}>{s.compliantPct}%</div>
          <div className="flex-1">
            <div className="text-sm font-medium">{t('compliance.complianceScore')}</div>
            <div className="mt-0.5 text-[11px] text-muted-foreground">
              {t('compliance.generatedAt', { date: df.formatDateTime(report.generatedAt) })}
              {' · '}
              {/* The window is what makes "12 alerts" mean anything. */}
              {t('compliance.window', {
                from: df.formatDate(report.windowFrom),
                to: df.formatDate(report.windowTo),
              })}
            </div>
            <div className="mt-2 h-1.5 overflow-hidden rounded-full bg-muted">
              <div className={cn('h-full rounded-full', scoreBg(s.compliantPct))} style={{ width: `${s.compliantPct}%` }} />
            </div>
            <div className="mt-1.5 text-[10px] text-muted-foreground">
              {t('compliance.scoredOver', { compliant: s.compliant, evaluated: s.evaluated, total: s.total })}
            </div>
          </div>
        </div>
        <div className="mt-4 grid grid-cols-3 gap-2 sm:grid-cols-7">
          <Stat label={t('compliance.status.COMPLIANT')} value={s.compliant} tone="text-emerald-500" />
          <Stat label={t('compliance.status.NON_COMPLIANT')} value={s.nonCompliant} tone="text-red-500" />
          <Stat label={t('compliance.status.AT_RISK')} value={s.atRisk} tone="text-amber-500" />
          <Stat label={t('compliance.status.NOT_COVERED')} value={s.notCovered} tone="text-zinc-400" />
          <Stat label={t('compliance.status.NOT_EVALUATED')} value={s.notEvaluated} tone="text-slate-400" />
          <Stat label={t('compliance.status.PENDING')} value={s.pending} tone="text-sky-400" />
          <Stat label={t('compliance.status.OUT_OF_SCOPE')} value={s.outOfScope} tone="text-violet-400" />
        </div>
      </div>

      {(report.sections ?? []).map((sec, i) => (
        <div key={sec.key || i} className="rounded-xl border border-border bg-card">
          <div className="flex items-center justify-between gap-3 border-b border-border px-4 py-2.5">
            <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">{sec.name}</span>
            {sec.summary && sec.summary.evaluated > 0 && (
              <span className={cn('text-xs font-semibold tabular-nums', scoreTone(sec.summary.compliantPct))}>
                {sec.summary.compliantPct}%
              </span>
            )}
          </div>
          <div>
            {(sec.requirements ?? []).map((req) => (
              <div key={req.id} className="border-b border-border last:border-0">
                <div className="flex items-baseline gap-2 bg-muted/30 px-4 py-1.5">
                  <span className={cn('rounded px-1.5 py-0.5 text-[10px] font-semibold', STATUS_TONE[req.status])}>
                    {t(`compliance.status.${req.status}`)}
                  </span>
                  <span className="font-mono text-[10px] text-muted-foreground">{req.id}</span>
                  <span className="truncate text-[12px]">{req.name}</span>
                </div>
                {(req.controlIds ?? []).map((cid) => {
                  const c = byId.get(cid)
                  if (!c) return null
                  return (
                    <ControlLine
                      key={`${req.id}:${cid}`}
                      report={report}
                      row={c}
                      onClick={onControlClick}
                      onChanged={onChanged}
                    />
                  )
                })}
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}

function ControlLine({
  report,
  row,
  onClick,
  onChanged,
}: {
  report: Report
  row: ControlRow
  onClick?: (row: ControlRow) => void
  onChanged?: (updated: Report) => void
}) {
  const { t } = useTranslation()
  return (
    <div
      onClick={onClick ? () => onClick(row) : undefined}
      className={cn(
        'flex items-start gap-3 px-4 py-2.5 pl-8',
        onClick && 'cursor-pointer transition-colors hover:bg-muted/50',
      )}
    >
      <ControlStatusBadge frameworkKey={report.frameworkKey} row={row} onChanged={onChanged} className="mt-0.5" />
      <div className="min-w-0 flex-1">
        <div className="flex items-baseline gap-2">
          <span className="font-mono text-[11px] text-muted-foreground">{row.controlId}</span>
          <span className="truncate text-[13px]">{row.name}</span>
        </div>
        {row.evidence && <p className="mt-0.5 text-[11px] text-muted-foreground">{row.evidence}</p>}
        {row.note && (
          <p
            className={cn(
              'mt-1 border-l-2 pl-2 text-[11px] italic text-muted-foreground',
              isOverridden(row) ? 'border-amber-500/50' : 'border-border',
            )}
          >
            “{row.note}”
            {row.editedBy && <span className="not-italic"> — {row.editedBy}</span>}
          </p>
        )}
        {/* The engine has changed its mind since this override was written. It
            still stands; only a person should withdraw it. */}
        {isEditStale(row) && (
          <p className="mt-1 flex items-center gap-1 text-[11px] text-amber-500">
            <AlertTriangle size={12} />
            {t('compliance.editStale', {
              was: t(`compliance.status.${row.originalStatus}`),
              now: t(`compliance.status.${row.engineStatus}`),
            })}
          </p>
        )}
      </div>
      <ControlNoteButton frameworkKey={report.frameworkKey} row={row} onSaved={onChanged} className="mt-0.5 shrink-0" />
      <div className="shrink-0 text-right text-[10px] text-muted-foreground">
        <div>{t('compliance.coverage', { n: row.coverage })}</div>
        <div>{t('compliance.activity', { n: row.activity })}</div>
      </div>
    </div>
  )
}

function Stat({ label, value, tone }: { label: string; value: number; tone: string }) {
  return (
    <div className="rounded-md border border-border bg-card px-2 py-1.5 text-center">
      <div className={cn('text-lg font-bold tabular-nums', tone)}>{value}</div>
      <div className="text-[9px] uppercase tracking-wide text-muted-foreground">{label}</div>
    </div>
  )
}
