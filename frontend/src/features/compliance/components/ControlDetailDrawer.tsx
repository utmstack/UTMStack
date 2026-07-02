import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, Loader2, X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { complianceService } from '../services/compliance-http.service'
import type { Control, ReportControlRow } from '../types/compliance.types'
import { STATUS_TONE } from './ReportView'

/** Right-side drawer showing full details of a report control row + its library Control. */
export function ControlDetailDrawer({ row, onClose }: { row: ReportControlRow; onClose: () => void }) {
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
              <span className={cn('rounded px-1.5 py-0.5 text-[10px] font-semibold', STATUS_TONE[row.status])}>
                {t(`compliance.status.${row.status}`)}
              </span>
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
            <Stat label={t('compliance.status.label', { defaultValue: 'Status' })} value={t(`compliance.status.${row.status}`)} />
            <Stat label={t('compliance.coverageLabel', { defaultValue: 'Coverage' })} value={t('compliance.coverage', { n: row.coverage })} />
            <Stat label={t('compliance.activityLabel', { defaultValue: 'Activity' })} value={t('compliance.activity', { n: row.activity })} />
            <Stat label={t('compliance.evidenceLabel', { defaultValue: 'Evidence' })} value={row.evidence || '—'} />
          </section>

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
                        {c.indexPattern && <div className="mt-1 font-mono text-[10px] text-muted-foreground">{c.indexPattern}</div>}
                        {c.sql && <pre className="mt-1.5 overflow-x-auto rounded bg-muted/60 p-2 font-mono text-[10px] leading-snug">{c.sql}</pre>}
                        {c.rule && (
                          <div className="mt-1.5 text-[11px] text-muted-foreground">
                            {c.rule}
                            {c.ruleValue != null && ` = ${c.ruleValue}`}
                            {c.field && ` · ${c.field}`}
                            {c.expected != null && ` = ${c.expected}`}
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
