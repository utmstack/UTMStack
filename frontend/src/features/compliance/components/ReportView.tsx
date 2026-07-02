import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { useDateFormat } from '@/shared/lib/datetime'
import type { ControlStatus, Report, ReportControlRow } from '../types/compliance.types'

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

export const STATUS_TONE: Record<ControlStatus, string> = {
  COMPLIANT: 'bg-emerald-500/15 text-emerald-500',
  NON_COMPLIANT: 'bg-red-500/15 text-red-500',
  AT_RISK: 'bg-amber-500/15 text-amber-500',
  NOT_COVERED: 'bg-zinc-500/15 text-zinc-400',
  OUT_OF_SCOPE: 'bg-violet-500/15 text-violet-400',
  PENDING: 'bg-sky-500/15 text-sky-400',
}

/** Renders a compliance report: summary score + per-section control breakdown. */
export function ReportView({ report, onControlClick }: { report: Report; onControlClick?: (row: ReportControlRow) => void }) {
  const { t } = useTranslation()
  const df = useDateFormat()
  const s = report.summary ?? { compliantPct: 0, total: 0, compliant: 0, nonCompliant: 0, atRisk: 0, notCovered: 0, pending: 0, outOfScope: 0 }
  return (
    <div className="space-y-5">
      <div className="rounded-xl border border-border bg-background/40 p-5">
        <div className="flex items-center gap-5">
          <div className={cn('text-4xl font-bold tabular-nums', scoreTone(s.compliantPct))}>{s.compliantPct}%</div>
          <div className="flex-1">
            <div className="text-sm font-medium">{t('compliance.complianceScore')}</div>
            <div className="mt-0.5 text-[11px] text-muted-foreground">{t('compliance.generatedAt', { date: df.formatDateTime(report.generatedAt) })}</div>
            <div className="mt-2 h-1.5 overflow-hidden rounded-full bg-muted">
              <div className={cn('h-full rounded-full', scoreBg(s.compliantPct))} style={{ width: `${s.compliantPct}%` }} />
            </div>
          </div>
        </div>
        <div className="mt-4 grid grid-cols-3 gap-2 sm:grid-cols-6">
          <Stat label={t('compliance.status.COMPLIANT')} value={s.compliant} tone="text-emerald-500" />
          <Stat label={t('compliance.status.NON_COMPLIANT')} value={s.nonCompliant} tone="text-red-500" />
          <Stat label={t('compliance.status.AT_RISK')} value={s.atRisk} tone="text-amber-500" />
          <Stat label={t('compliance.status.NOT_COVERED')} value={s.notCovered} tone="text-zinc-400" />
          <Stat label={t('compliance.status.PENDING')} value={s.pending} tone="text-sky-400" />
          <Stat label={t('compliance.status.OUT_OF_SCOPE')} value={s.outOfScope} tone="text-violet-400" />
        </div>
      </div>

      {(report.sections ?? []).map((sec, i) => (
        <div key={i} className="rounded-xl border border-border bg-card">
          <div className="border-b border-border px-4 py-2.5 text-xs font-semibold uppercase tracking-wider text-muted-foreground">{sec.name}</div>
          <div>
            {(sec.controls ?? []).map((c) => (
              <div
                key={c.controlId}
                onClick={onControlClick ? () => onControlClick(c) : undefined}
                className={cn(
                  'flex items-start gap-3 border-b border-border px-4 py-2.5 last:border-0',
                  onControlClick && 'cursor-pointer transition-colors hover:bg-muted/50',
                )}
              >
                <span className={cn('mt-0.5 shrink-0 rounded px-1.5 py-0.5 text-[10px] font-semibold', STATUS_TONE[c.status])}>
                  {t(`compliance.status.${c.status}`)}
                </span>
                <div className="min-w-0 flex-1">
                  <div className="flex items-baseline gap-2">
                    <span className="font-mono text-[11px] text-muted-foreground">{c.controlId}</span>
                    <span className="truncate text-[13px]">{c.name}</span>
                  </div>
                  {c.evidence && <p className="mt-0.5 text-[11px] text-muted-foreground">{c.evidence}</p>}
                </div>
                <div className="shrink-0 text-right text-[10px] text-muted-foreground">
                  <div>{t('compliance.coverage', { n: c.coverage })}</div>
                  <div>{t('compliance.activity', { n: c.activity })}</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      ))}
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
