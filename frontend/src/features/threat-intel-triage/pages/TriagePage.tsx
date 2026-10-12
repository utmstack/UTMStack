import { useTranslation } from 'react-i18next'
import { Download, Play, RotateCcw, Square } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { Textarea } from '@/shared/components/ui/textarea'
import { useTriage } from '../hooks/use-triage'
import { MAX_TRIAGE_ITEMS, toStix, type Verdict } from '../domain/triage.types'
import { toCsv } from '../services/csv'
import { useTiConfigStatus } from '@/features/threat-intel/hooks/use-ti-config-status'
import { NotConfiguredState } from '@/features/threat-intel/components/NotConfiguredState'
import { cn } from '@/shared/lib/utils'

const VERDICT_ORDER: Verdict[] = ['malicious', 'suspicious', 'unknown', 'benign', 'notFound', 'invalid']
const VERDICT_LABEL: Record<Verdict, string> = {
  malicious: 'threatIntel.triage.v.malicious',
  suspicious: 'threatIntel.triage.v.suspicious',
  benign: 'threatIntel.triage.v.benign',
  unknown: 'threatIntel.triage.v.unknown',
  notFound: 'threatIntel.triage.v.notFound',
  invalid: 'threatIntel.triage.v.invalid',
}

function verdictClass(verdict: Verdict): string {
  switch (verdict) {
    case 'malicious':
      return 'border-red-300 bg-red-50 text-red-700'
    case 'suspicious':
      return 'border-amber-300 bg-amber-50 text-amber-700'
    case 'benign':
      return 'border-emerald-300 bg-emerald-50 text-emerald-700'
    default:
      return 'border-border/50 text-muted-foreground'
  }
}

/**
 * Bulk indicator triage.
 *
 * An incident arrives as a block of indicators, not one lookup at a time.
 * This resolves the whole batch, ranks it worst-first, and exports the result.
 */
export function TriagePage() {
  const { t } = useTranslation()
  const { isConfigured } = useTiConfigStatus()
  const { raw, setRaw, parsed, sorted, counts, running, done, total, run, stop, reset } = useTriage()

  if (!isConfigured) return <NotConfiguredState />

  const download = (content: string, filename: string, mime: string) => {
    const url = URL.createObjectURL(new Blob([content], { type: mime }))
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    link.click()
    // Revoking immediately would race the download in some browsers.
    setTimeout(() => URL.revokeObjectURL(url), 10_000)
  }

  return (
    <div className="w-full px-6 pb-6 pt-3">
      <div className="mx-auto max-w-6xl space-y-4">
        <header className="flex flex-wrap items-baseline gap-3">
          <h1 className="text-xl font-semibold">{t('threatIntel.triage.title')}</h1>
          <span className="text-xs text-muted-foreground">{t('threatIntel.triage.subtitle')}</span>
        </header>

        <section className="rounded-xl border border-border bg-card p-4 space-y-2">
          <label
            htmlFor="triage-input"
            className="block text-[11px] uppercase tracking-wider text-muted-foreground"
          >
            {t('threatIntel.triage.input')}
          </label>
          <Textarea
            id="triage-input"
            name="indicators"
            value={raw}
            onChange={e => setRaw(e.target.value)}
            rows={6}
            spellCheck={false}
            dir="ltr"
            placeholder={t('threatIntel.triage.placeholder')}
            className="font-mono text-xs"
          />

          <div className="flex flex-wrap items-center gap-2">
            <span className="text-xs text-muted-foreground">
              {t('threatIntel.triage.parsed', { count: parsed.length })}
            </span>
            {parsed.length > MAX_TRIAGE_ITEMS && (
              <span className="text-xs text-amber-600">
                {t('threatIntel.triage.capped', { max: MAX_TRIAGE_ITEMS })}
              </span>
            )}
            {running && (
              <span className="text-xs tabular-nums text-primary">
                {t('threatIntel.triage.progress', { done, total })}
              </span>
            )}

            <div className="ms-auto flex items-center gap-2">
              {running ? (
                <Button variant="outline" size="sm" onClick={stop}>
                  <Square size={14} className="mr-1.5" />
                  {t('threatIntel.triage.stop')}
                </Button>
              ) : (
                <Button size="sm" onClick={run} disabled={parsed.length === 0}>
                  <Play size={14} className="mr-1.5" />
                  {t('threatIntel.triage.run')}
                </Button>
              )}
              <Button variant="outline" size="sm" onClick={reset} disabled={!raw && sorted.length === 0}>
                <RotateCcw size={14} className="mr-1.5" />
                {t('threatIntel.triage.reset')}
              </Button>
            </div>
          </div>
        </section>

        {sorted.length > 0 && (
          <section className="rounded-xl border border-border bg-card p-4 space-y-3">
            <div className="flex flex-wrap items-center gap-2">
              <h2 className="text-[11px] uppercase tracking-wider text-muted-foreground">
                {t('threatIntel.triage.summary')}
              </h2>
              {VERDICT_ORDER.filter(v => counts[v]).map(v => (
                <span
                  key={v}
                  className={cn(
                    'rounded-sm border px-1.5 py-0.5 font-mono text-[10px] tabular-nums',
                    verdictClass(v),
                  )}
                >
                  {t(VERDICT_LABEL[v])} {counts[v]}
                </span>
              ))}

              <div className="ms-auto flex items-center gap-2">
                <Button variant="outline" size="sm" onClick={() => download(toCsv(sorted), 'triage.csv', 'text/csv;charset=utf-8')}>
                  <Download size={14} className="mr-1.5" />
                  {t('threatIntel.triage.exportCsv')}
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  title={t('threatIntel.triage.stixNote')}
                  onClick={() => download(toStix(sorted), 'triage.stix.json', 'application/json')}
                >
                  <Download size={14} className="mr-1.5" />
                  {t('threatIntel.triage.exportStix')}
                </Button>
              </div>
            </div>

            {/* Wide table scrolls inside its own container so the page never
                scrolls horizontally. */}
            <div className="overflow-x-auto">
              <table className="w-full min-w-[46rem] border-collapse text-xs">
                <thead>
                  <tr className="border-b border-border/40 text-muted-foreground">
                    <Th>{t('threatIntel.triage.indicator')}</Th>
                    <Th>{t('threatIntel.triage.type')}</Th>
                    <Th>{t('threatIntel.triage.verdict')}</Th>
                    <Th className="text-right">{t('threatIntel.triage.reputation')}</Th>
                    <Th className="text-right">{t('threatIntel.triage.accuracy')}</Th>
                    <Th>{t('threatIntel.triage.tags')}</Th>
                    <Th>{t('threatIntel.triage.lastSeen')}</Th>
                    <Th />
                  </tr>
                </thead>
                <tbody>
                  {sorted.map(row => (
                    <tr key={row.key} className="border-b border-border/20 hover:bg-muted/30">
                      <td className="max-w-[18rem] truncate py-1.5 pr-2" dir="ltr">
                        {row.input}
                      </td>
                      <td className="py-1.5 pr-2 uppercase text-muted-foreground">{row.type || '—'}</td>
                      <td className="py-1.5 pr-2">
                        <span className={cn('rounded-sm border px-1.5 py-0.5 font-mono text-[10px]', verdictClass(row.verdict))}>
                          {t(VERDICT_LABEL[row.verdict])}
                        </span>
                      </td>
                      <td className={cn('py-1.5 pr-2 text-right tabular-nums', (row.entity?.reputation ?? 0) < 0 ? 'text-red-600' : 'text-muted-foreground')}>
                        {row.entity?.reputation ?? '—'}
                      </td>
                      <td className="py-1.5 pr-2 text-right tabular-nums text-muted-foreground">
                        {row.entity?.accuracy ?? '—'}
                      </td>
                      <td className="max-w-[12rem] truncate py-1.5 pr-2 text-muted-foreground">
                        {row.entity?.tags.join(' ') || '—'}
                      </td>
                      <td className="py-1.5 pr-2 text-muted-foreground/70">
                        {row.entity?.lastSeen.slice(0, 10) || '—'}
                      </td>
                      <td />
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </section>
        )}

        {sorted.length === 0 && !running && (
          <p className="py-12 text-center text-xs text-muted-foreground">{t('threatIntel.triage.empty')}</p>
        )}
      </div>
    </div>
  )
}

function Th({ children, className }: { children?: React.ReactNode; className?: string }) {
  return (
    <th scope="col" className={cn('py-1.5 pr-2 text-left font-normal uppercase tracking-widest text-[9px]', className)}>
      {children}
    </th>
  )
}
