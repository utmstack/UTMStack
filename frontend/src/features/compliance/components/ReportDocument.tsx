import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Download, Loader2, RefreshCw, X } from 'lucide-react'
import { toast } from 'sonner'
import { useDateFormat } from '@/shared/lib/datetime'
import { useBranding } from '@/features/branding'
import type { ControlStatus, Report, ReportControlRow, ReportSection } from '../types/compliance.types'

const STATUS_HEX: Record<ControlStatus, string> = {
  COMPLIANT: '#10b981',
  NON_COMPLIANT: '#ef4444',
  AT_RISK: '#f59e0b',
  NOT_COVERED: '#a1a1aa',
  OUT_OF_SCOPE: '#a78bfa',
  PENDING: '#38bdf8',
}
const ORDER: ControlStatus[] = ['COMPLIANT', 'NON_COMPLIANT', 'AT_RISK', 'NOT_COVERED', 'PENDING', 'OUT_OF_SCOPE']

function scoreHex(score: number): string {
  if (score >= 80) return '#10b981'
  if (score >= 50) return '#f59e0b'
  return '#ef4444'
}

/** Professional, print-ready compliance report rendered as a paper document. */
export function ReportDocument({
  report,
  onClose,
  onRun,
  running,
  onDownloadPdf,
}: {
  report: Report
  onClose: () => void
  onRun?: () => void
  running?: boolean
  onDownloadPdf?: () => Promise<Blob>
}) {
  const { t } = useTranslation()
  const df = useDateFormat()
  const { branding } = useBranding()
  const [downloading, setDownloading] = useState(false)

  const download = async () => {
    if (!onDownloadPdf || downloading) return
    setDownloading(true)
    try {
      const blob = await onDownloadPdf()
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${(report.frameworkName || 'Compliance_Report').replace(/[^\w -]/g, '').trim() || 'Compliance_Report'}.pdf`
      document.body.appendChild(a)
      a.click()
      a.remove()
      URL.revokeObjectURL(url)
    } catch {
      toast.error(t('compliance.doc.downloadError'))
    } finally {
      setDownloading(false)
    }
  }
  // White-label when configured; otherwise UTMStack defaults.
  const brandName = (branding?.enabled && branding.productName) || 'UTMStack'
  const brandLogo = (branding?.enabled && branding.logoUrl) || '/logo.svg'
  const s = report.summary ?? { compliantPct: 0, total: 0, compliant: 0, nonCompliant: 0, atRisk: 0, notCovered: 0, pending: 0, outOfScope: 0 }
  const sections = report.sections ?? []

  const counts: Record<ControlStatus, number> = {
    COMPLIANT: s.compliant,
    NON_COMPLIANT: s.nonCompliant,
    AT_RISK: s.atRisk,
    NOT_COVERED: s.notCovered,
    PENDING: s.pending,
    OUT_OF_SCOPE: s.outOfScope,
  }
  const totalAll = ORDER.reduce((a, k) => a + counts[k], 0)
  const statusLabel = s.compliantPct >= 80 ? t('compliance.status.COMPLIANT') : s.compliantPct >= 50 ? t('compliance.status.AT_RISK') : t('compliance.status.NON_COMPLIANT')

  return (
    <div className="fixed inset-0 z-[60] flex flex-col bg-black/70 backdrop-blur-sm">
      {/* Toolbar (not printed) */}
      <div className="flex shrink-0 items-center justify-between gap-4 px-4 py-2.5 text-white print:hidden">
        <span className="truncate text-sm font-medium">{report.frameworkName} · {t('compliance.doc.title')}</span>
        <div className="flex items-center gap-2">
          {onRun && (
            <button onClick={onRun} disabled={running} className="inline-flex items-center gap-1.5 rounded-md bg-white/10 px-3 py-1.5 text-xs hover:bg-white/20 disabled:opacity-50">
              {running ? <Loader2 size={13} className="animate-spin" /> : <RefreshCw size={13} />} {t('compliance.runReport')}
            </button>
          )}
          {onDownloadPdf && (
            <button onClick={() => void download()} disabled={downloading} className="inline-flex items-center gap-1.5 rounded-md bg-white px-3 py-1.5 text-xs font-medium text-zinc-900 hover:bg-zinc-100 disabled:opacity-60">
              {downloading ? <Loader2 size={13} className="animate-spin" /> : <Download size={13} />} {t('compliance.doc.download')}
            </button>
          )}
          <button onClick={onClose} className="flex h-8 w-8 items-center justify-center rounded-md hover:bg-white/10"><X size={18} /></button>
        </div>
      </div>

      {/* Scrollable document */}
      <div className="flex-1 overflow-y-auto px-4 pb-12">
        <div id="report-doc" className="mx-auto my-4 max-w-[820px] rounded-sm bg-white text-zinc-800 shadow-2xl">
          {/* ── Cover ─────────────────────────────────────────────── */}
          <section className="flex min-h-[60vh] flex-col items-center justify-center border-b border-zinc-200 px-16 py-20 text-center">
            <div className="mb-8 flex items-center gap-2.5">
              <img src={brandLogo} alt={brandName} className="h-9 w-auto object-contain" onError={(e) => { e.currentTarget.style.display = 'none' }} />
              <span className="text-sm font-semibold uppercase tracking-[0.2em] text-zinc-500">{brandName}</span>
            </div>
            <div className="text-xs font-semibold uppercase tracking-[0.3em] text-zinc-400">{t('compliance.doc.title')}</div>
            <h1 className="mt-3 text-4xl font-bold leading-tight text-zinc-900">{report.frameworkName}</h1>
            <div className="my-10">
              <ScoreGauge score={s.compliantPct} />
            </div>
            <div className="inline-flex items-center gap-2 rounded-full px-4 py-1.5 text-sm font-semibold" style={{ backgroundColor: `${scoreHex(s.compliantPct)}1a`, color: scoreHex(s.compliantPct) }}>
              {statusLabel}
            </div>
            <div className="mt-10 text-xs text-zinc-500">
              <div>{t('compliance.doc.generated', { date: df.formatDateTime(report.generatedAt) })}</div>
              <div className="mt-1">{t('compliance.doc.controlsEvaluated', { n: totalAll })}</div>
            </div>
          </section>

          {/* ── Executive summary ─────────────────────────────────── */}
          <section className="px-16 py-12">
            <SectionHeading n="01" title={t('compliance.doc.executiveSummary')} />
            <div className="mt-6 flex flex-wrap items-center gap-10">
              <StatusDonut counts={counts} total={totalAll} />
              <div className="min-w-[260px] flex-1">
                <table className="w-full text-sm">
                  <tbody>
                    {ORDER.map((k) => (
                      <tr key={k} className="border-b border-zinc-100 last:border-0">
                        <td className="py-2">
                          <span className="mr-2 inline-block h-2.5 w-2.5 rounded-full align-middle" style={{ backgroundColor: STATUS_HEX[k] }} />
                          <span className="align-middle text-zinc-700">{t(`compliance.status.${k}`)}</span>
                        </td>
                        <td className="py-2 text-right font-semibold tabular-nums text-zinc-900">{counts[k]}</td>
                        <td className="w-14 py-2 text-right text-xs tabular-nums text-zinc-400">{totalAll ? Math.round((counts[k] / totalAll) * 100) : 0}%</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </section>

          {/* ── Per-section breakdown ─────────────────────────────── */}
          <section className="px-16 pb-16">
            <SectionHeading n="02" title={t('compliance.doc.bySection')} />
            <div className="mt-6 space-y-8">
              {sections.map((sec, i) => (
                <SectionBlock key={i} section={sec} t={t} />
              ))}
            </div>
          </section>

          <footer className="border-t border-zinc-200 px-16 py-5 text-center text-[10px] text-zinc-400">
            {t('compliance.doc.footer', { framework: report.frameworkName, date: df.formatDateTime(report.generatedAt), brand: brandName })}
          </footer>
        </div>
      </div>
    </div>
  )
}

function SectionHeading({ n, title }: { n: string; title: string }) {
  return (
    <div className="flex items-baseline gap-3 border-b-2 border-zinc-900 pb-2">
      <span className="text-xs font-bold tabular-nums text-zinc-300">{n}</span>
      <h2 className="text-lg font-bold uppercase tracking-wide text-zinc-900">{title}</h2>
    </div>
  )
}

function sectionPct(controls: ReportControlRow[]): number {
  const rel = controls.filter((c) => c.status === 'COMPLIANT' || c.status === 'NON_COMPLIANT' || c.status === 'AT_RISK')
  if (rel.length === 0) return 0
  return Math.round((rel.filter((c) => c.status === 'COMPLIANT').length / rel.length) * 100)
}

function SectionBlock({ section, t }: { section: ReportSection; t: ReturnType<typeof useTranslation>['t'] }) {
  const controls = section.controls ?? []
  const pct = sectionPct(controls)
  return (
    <div className="break-inside-avoid">
      <div className="mb-2 flex items-center justify-between gap-4">
        <h3 className="text-sm font-semibold text-zinc-800">{section.name}</h3>
        <div className="flex items-center gap-2">
          <div className="h-1.5 w-32 overflow-hidden rounded-full bg-zinc-100">
            <div className="h-full rounded-full" style={{ width: `${pct}%`, backgroundColor: scoreHex(pct) }} />
          </div>
          <span className="w-9 text-right text-xs font-semibold tabular-nums text-zinc-600">{pct}%</span>
        </div>
      </div>
      <table className="w-full border-collapse text-[12px]">
        <thead>
          <tr className="border-b border-zinc-200 text-left text-[10px] uppercase tracking-wide text-zinc-400">
            <th className="py-1.5 pr-2 font-medium">{t('compliance.doc.controlCol')}</th>
            <th className="py-1.5 pr-2 font-medium">{t('compliance.doc.statusCol')}</th>
            <th className="py-1.5 pr-2 font-medium">{t('compliance.doc.evidenceCol')}</th>
            <th className="py-1.5 text-right font-medium">{t('compliance.doc.coverageCol')}</th>
          </tr>
        </thead>
        <tbody>
          {controls.map((c) => (
            <tr key={c.controlId} className="border-b border-zinc-100 align-top last:border-0">
              <td className="py-2 pr-2">
                <div className="font-mono text-[10px] text-zinc-400">{c.controlId}</div>
                <div className="text-zinc-800">{c.name}</div>
              </td>
              <td className="py-2 pr-2">
                <span className="inline-block rounded px-1.5 py-0.5 text-[10px] font-semibold" style={{ backgroundColor: `${STATUS_HEX[c.status]}22`, color: STATUS_HEX[c.status] }}>
                  {t(`compliance.status.${c.status}`)}
                </span>
              </td>
              <td className="py-2 pr-2 text-[11px] text-zinc-500">{c.evidence || '—'}</td>
              <td className="py-2 text-right text-[11px] tabular-nums text-zinc-500">{c.coverage} / {c.activity}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

/* ── SVG charts (no deps) ─────────────────────────────────────────────── */

function ScoreGauge({ score }: { score: number }) {
  const r = 70
  const c = 2 * Math.PI * r
  const dash = (Math.max(0, Math.min(100, score)) / 100) * c
  const color = scoreHex(score)
  return (
    <svg width="180" height="180" viewBox="0 0 180 180">
      <circle cx="90" cy="90" r={r} fill="none" stroke="#f1f1f4" strokeWidth="14" />
      <circle cx="90" cy="90" r={r} fill="none" stroke={color} strokeWidth="14" strokeLinecap="round" strokeDasharray={`${dash} ${c - dash}`} transform="rotate(-90 90 90)" />
      <text x="90" y="86" textAnchor="middle" fontSize="40" fontWeight="700" fill="#18181b">{score}%</text>
      <text x="90" y="108" textAnchor="middle" fontSize="11" fill="#a1a1aa" style={{ textTransform: 'uppercase', letterSpacing: '0.1em' }}>score</text>
    </svg>
  )
}

function StatusDonut({ counts, total }: { counts: Record<ControlStatus, number>; total: number }) {
  const r = 58
  const c = 2 * Math.PI * r
  let offset = 0
  const segs =
    total === 0
      ? [{ color: '#e4e4e7', len: c, off: 0 }]
      : ORDER.filter((k) => counts[k] > 0).map((k) => {
          const len = (counts[k] / total) * c
          const seg = { color: STATUS_HEX[k], len, off: offset }
          offset += len
          return seg
        })
  return (
    <svg width="150" height="150" viewBox="0 0 150 150" className="shrink-0">
      <circle cx="75" cy="75" r={r} fill="none" stroke="#f1f1f4" strokeWidth="18" />
      {segs.map((sg, i) => (
        <circle key={i} cx="75" cy="75" r={r} fill="none" stroke={sg.color} strokeWidth="18" strokeDasharray={`${sg.len} ${c - sg.len}`} strokeDashoffset={-sg.off} transform="rotate(-90 75 75)" />
      ))}
      <text x="75" y="72" textAnchor="middle" fontSize="26" fontWeight="700" fill="#18181b">{total}</text>
      <text x="75" y="90" textAnchor="middle" fontSize="9" fill="#a1a1aa" style={{ textTransform: 'uppercase', letterSpacing: '0.1em' }}>controls</text>
    </svg>
  )
}
