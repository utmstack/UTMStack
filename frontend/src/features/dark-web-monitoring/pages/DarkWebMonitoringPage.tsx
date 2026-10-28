import { useCallback, useEffect, useMemo, useRef, useState, type UIEvent } from 'react'
import { CalendarDays, ChevronDown, Download, Eye, FileText, Loader2, Search, SlidersHorizontal } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { FloatingPanel } from '@/shared/components/ui/floating-panel'
import { AdvancedExportOptionsDialog } from '../components/AdvancedExportOptionsDialog'
import { BreachDetailsDrawer } from '../components/BreachDetailsDrawer'
import { StatusPill } from '../components/StatusPill'
import { CustomRangeExportDialog } from '../components/CustomRangeExportDialog'
import { TimeRangeFilter, type TimeFilter } from '../components/TimeRangeFilter'
import type { ExportAction } from '../components/export-types'
import { useDarkWebData } from '../hooks/use-dark-web-data'
import { MASKED, buildExportRows, downloadCsv, printReportPdf } from '../lib/export'
import { EMPTY_VALUE, STATUS_KEYS, formatDateShort, valueOrDash } from '../lib/format'
import { SURFACE_OF, filterRecords, paginateRecords } from '../lib/record-query'
import type { DarkWebRecord, LeakSurface } from '../types'

type StatusFilter = 'all' | 'OPEN' | 'IN_MITIGATION' | 'COMPLETED'

const PAGE_SIZE = 5
const SURFACES: LeakSurface[] = ['domains', 'emails', 'ips', 'usernames', 'phones']
const STATUSES: StatusFilter[] = ['all', 'OPEN', 'IN_MITIGATION', 'COMPLETED']

function getCredential(record: DarkWebRecord): string | null {
  return record.password ?? record.hashedPassword
}

export function DarkWebMonitoringPage() {
  const { t, i18n } = useTranslation()
  const { records, stats, loading, error, retry, setRecordStatus } = useDarkWebData()
  const [surface, setSurface] = useState<LeakSurface>('domains')
  const [query, setQuery] = useState('')
  const [status, setStatus] = useState<StatusFilter>('all')
  const [time, setTime] = useState<TimeFilter>({ kind: 'preset', preset: 'all' })
  const [employeesOnly, setEmployeesOnly] = useState(false)
  const [reveal, setReveal] = useState(false)
  const [visibleCount, setVisibleCount] = useState(PAGE_SIZE)
  const [loadingMore, setLoadingMore] = useState(false)
  const loadMoreTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const tableScrollRef = useRef<HTMLDivElement>(null)
  const [selectedRecordId, setSelectedRecordId] = useState<string | null>(null)
  const [reportMenuOpen, setReportMenuOpen] = useState(false)
  const [customRangeOpen, setCustomRangeOpen] = useState(false)
  const [advancedExportOpen, setAdvancedExportOpen] = useState(false)
  const reportButtonRef = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    setVisibleCount(PAGE_SIZE)
    setLoadingMore(false)
    if (loadMoreTimer.current) clearTimeout(loadMoreTimer.current)
    return () => {
      if (loadMoreTimer.current) clearTimeout(loadMoreTimer.current)
    }
  }, [employeesOnly, query, status, surface, time])

  const filtered = useMemo(
    () => filterRecords(records, { surface, status, employeesOnly, search: query, time }),
    [records, employeesOnly, query, status, surface, time],
  )

  const visibleRecords = useMemo(() => paginateRecords(filtered, visibleCount), [filtered, visibleCount])
  const selectedRecord = selectedRecordId ? records.find((record) => record.id === selectedRecordId) : undefined
  const hasMoreRecords = visibleCount < filtered.length

  const surfaceCounts = useMemo(() => {
    const counts: Record<LeakSurface, number> = { domains: 0, emails: 0, ips: 0, usernames: 0, phones: 0 }
    for (const record of records) counts[SURFACE_OF[record.organizationElementType]] += 1
    // The organization endpoint owns the headline numbers; fall back to the
    // fetched records until it exists.
    return {
      domains: stats?.totalDomainBreaches ?? counts.domains,
      emails: stats?.totalEmailsBreaches ?? counts.emails,
      ips: stats?.totalIpsBreaches ?? counts.ips,
      usernames: stats?.totalUsernamesBreaches ?? counts.usernames,
      phones: stats?.totalPhonesBreaches ?? counts.phones,
    }
  }, [records, stats])

  const statusCounts = useMemo(
    () => ({
      all: stats?.totalBreaches ?? records.length,
      OPEN: records.filter((record) => record.breachStatus === 'OPEN').length,
      IN_MITIGATION: records.filter((record) => record.breachStatus === 'IN_MITIGATION').length,
      COMPLETED: records.filter((record) => record.breachStatus === 'COMPLETED').length,
    }),
    [records, stats],
  )

  const employeeCount = useMemo(() => records.filter((record) => record.employee).length, [records])
  const monitoredDomains = useMemo(
    () =>
      Array.from(
        new Set(
          records
            .filter((record) => record.organizationElementType === 'DOMAIN')
            .map((record) => record.organizationElement),
        ),
      ),
    [records],
  )
  const exposedSurfaceCount = SURFACES.filter((key) => surfaceCounts[key] > 0).length
  const employeeShare = records.length > 0 ? (employeeCount / records.length) * 100 : 0

  const handleExport = () => downloadCsv('dark-web-monitoring-demo.csv', buildExportRows(filtered, t, !reveal, (value) => formatDateShort(value, i18n.language)))
  const handleReport = () =>
    downloadCsv('dark-web-monitoring-demo-report.csv', [
      [t('darkWebMonitoring.title'), t('darkWebMonitoring.demoNotice')],
      [t('darkWebMonitoring.summary.records'), String(statusCounts.all)],
      [t('darkWebMonitoring.summary.employeeCredentials'), String(employeeCount)],
      [t('darkWebMonitoring.summary.open'), String(statusCounts.OPEN)],
      [t('darkWebMonitoring.summary.mitigation'), String(statusCounts.IN_MITIGATION)],
      [t('darkWebMonitoring.summary.completed'), String(statusCounts.COMPLETED)],
    ])
  const handleDialogExport = (items: DarkWebRecord[], action: ExportAction, filename: string, maskSensitive: boolean): boolean => {
    try {
      const rows = buildExportRows(items, t, maskSensitive, (value) => formatDateShort(value, i18n.language))
      if (action === 'CSV') {
        downloadCsv(`${filename}.csv`, rows)
        return true
      }
      // The print window doubles as the PDF report; if the browser blocks the
      // pop-up the report is not produced, so warn and keep the dialog open.
      if (!printReportPdf(t('darkWebMonitoring.title'), rows)) {
        toast.error(t('darkWebMonitoring.exportDialogs.exportFailed'))
        return false
      }
      return true
    } catch {
      toast.error(t('darkWebMonitoring.exportDialogs.exportFailed'))
      return false
    }
  }
  const openCustomRange = () => {
    setReportMenuOpen(false)
    setCustomRangeOpen(true)
  }
  const openAdvancedExport = () => {
    setReportMenuOpen(false)
    setAdvancedExportOpen(true)
  }
  const loadMore = useCallback(() => {
    if (loadingMore || !hasMoreRecords) return
    setLoadingMore(true)
    loadMoreTimer.current = setTimeout(() => {
      setVisibleCount((current) => Math.min(current + PAGE_SIZE, filtered.length))
      setLoadingMore(false)
      loadMoreTimer.current = null
    }, 300)
  }, [hasMoreRecords, loadingMore, filtered.length])

  // Scroll drives loading while the table overflows; when the visible rows
  // already fit the viewport there is no scrollbar, so keep loading until the
  // table fills or everything is shown.
  useEffect(() => {
    const el = tableScrollRef.current
    if (!el || el.scrollHeight <= 0 || el.scrollHeight > el.clientHeight) return
    loadMore()
  }, [filtered.length, visibleCount, loadMore])
  const handleTableScroll = (event: UIEvent<HTMLDivElement>) => {
    const element = event.currentTarget
    if (element.scrollHeight - element.scrollTop - element.clientHeight < 300) loadMore()
  }

  if (loading) {
    return (
      <div className="flex min-h-full w-full flex-col gap-4 px-6 pb-6 pt-3">
        <header className="flex flex-wrap items-start justify-between gap-4">
          <div className="h-6 w-44 animate-pulse rounded bg-muted" />
          <div className="h-8 w-28 animate-pulse rounded-md bg-muted" />
        </header>
        <div className="h-12 animate-pulse rounded-xl bg-muted/60" />
        <div className="grid gap-4 lg:grid-cols-2">
          <div className="h-40 animate-pulse rounded-xl bg-muted/60" />
          <div className="h-40 animate-pulse rounded-xl bg-muted/60" />
        </div>
        <div role="status" className="flex items-center justify-center gap-3 py-16 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" />
          {t('darkWebMonitoring.loading')}
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex min-h-full w-full flex-col items-center justify-center gap-3 px-6" role="alert">
        <p className="text-sm text-muted-foreground">{t('darkWebMonitoring.loadingFailed')}</p>
        <Button size="sm" variant="outline" onClick={retry}>
          {t('darkWebMonitoring.retry')}
        </Button>
      </div>
    )
  }

  return (
    <div className="flex min-h-full w-full flex-col gap-4 px-6 pb-6 pt-3">
      <header className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="text-xl font-semibold">{t('darkWebMonitoring.title')}</h1>
          <span className="text-xs text-muted-foreground">{t('darkWebMonitoring.subtitle')}</span>
        </div>
        <button
          ref={reportButtonRef}
          type="button"
          aria-haspopup="menu"
          aria-expanded={reportMenuOpen}
          onClick={() => setReportMenuOpen((current) => !current)}
          className="inline-flex h-8 items-center gap-2 rounded-md bg-primary px-3 text-xs font-medium text-primary-foreground transition-colors hover:bg-primary/90"
        >
          <FileText size={14} />
          {t('darkWebMonitoring.report')}
          <span className="ml-1 flex h-8 items-center border-l border-white/15 pl-2">
            <ChevronDown size={13} />
          </span>
        </button>
        <FloatingPanel
          open={reportMenuOpen}
          onClose={() => setReportMenuOpen(false)}
          triggerRef={reportButtonRef}
          align="end"
          className="z-[70] w-60 rounded-lg border border-border bg-popover p-2 shadow-xl"
        >
          <div role="menu">
            <div className="px-2 py-1.5">
              <div className="text-xs font-semibold">{t('darkWebMonitoring.reportMenu.title')}</div>
              <p className="mt-1 text-[10px] leading-relaxed text-muted-foreground">
                {t('darkWebMonitoring.reportMenu.description')}
              </p>
            </div>
            <div className="my-1 border-t border-border" />
            <button
              type="button"
              role="menuitem"
              onClick={() => {
                handleReport()
                setReportMenuOpen(false)
              }}
              className="flex w-full items-center gap-2 rounded-md px-2 py-2 text-left text-xs hover:bg-muted"
            >
              <Download size={14} className="shrink-0 text-muted-foreground" />
              {t('darkWebMonitoring.reportMenu.exportCsv')}
            </button>
            <button
              type="button"
              role="menuitem"
              onClick={openCustomRange}
              className="flex w-full items-center gap-2 rounded-md px-2 py-2 text-left text-xs hover:bg-muted"
            >
              <CalendarDays size={14} className="shrink-0 text-muted-foreground" />
              {t('darkWebMonitoring.reportMenu.customRange')}
            </button>
            <button
              type="button"
              role="menuitem"
              onClick={openAdvancedExport}
              className="flex w-full items-center gap-2 rounded-md px-2 py-2 text-left text-xs hover:bg-muted"
            >
              <SlidersHorizontal size={14} className="shrink-0 text-muted-foreground" />
              {t('darkWebMonitoring.reportMenu.advancedOptions')}
            </button>
          </div>
        </FloatingPanel>
      </header>

      {customRangeOpen && (
        <CustomRangeExportDialog
          records={filtered}
          onClose={() => setCustomRangeOpen(false)}
          onExport={handleDialogExport}
        />
      )}
      {advancedExportOpen && (
        <AdvancedExportOptionsDialog
          records={filtered}
          onClose={() => setAdvancedExportOpen(false)}
          onExport={handleDialogExport}
        />
      )}
      {selectedRecord && (
        <BreachDetailsDrawer
          record={selectedRecord}
          onClose={() => setSelectedRecordId(null)}
          onStatusChange={(nextStatus) => setRecordStatus(selectedRecord.id, nextStatus)}
        />
      )}

      <div className="rounded-xl border border-primary/20 bg-primary/5 px-4 py-2 text-xs text-primary">
        {t('darkWebMonitoring.demoNotice')}
      </div>

      <section className="grid overflow-hidden rounded-xl border border-border bg-card lg:grid-cols-[1fr_1.08fr]">
        <div className="border-b border-border p-4 lg:border-b-0 lg:border-r">
          <div className="text-[10px] font-semibold uppercase tracking-[0.18em] text-primary">
            {t('darkWebMonitoring.summary.employeeCredentials')}
          </div>
          <div className="mt-1 text-4xl font-semibold tabular-nums">{employeeCount}</div>
          <p className="mt-1 text-xs text-muted-foreground">
            {employeeShare.toFixed(1)}% {t('darkWebMonitoring.summary.ofRecords', { count: records.length })}
          </p>
          <div className="mt-3 h-2 overflow-hidden rounded-full bg-muted">
            <div className="h-full rounded-full bg-rose-500" style={{ width: `${employeeShare}%` }} />
          </div>
          <div className="mt-2 flex gap-4 text-[10px] text-muted-foreground">
            <span><span className="mr-1 inline-block h-1.5 w-1.5 rounded-full bg-rose-500" />{t('darkWebMonitoring.employee')} {employeeCount}</span>
            <span><span className="mr-1 inline-block h-1.5 w-1.5 rounded-full bg-muted-foreground/50" />{t('darkWebMonitoring.other')} {records.length - employeeCount}</span>
          </div>
        </div>
        <div className="grid grid-cols-2 gap-4 p-4 sm:gap-6">
          {([
            ['records', statusCounts.all, 'text-foreground'],
            ['open', statusCounts.OPEN, 'text-rose-500'],
            ['mitigation', statusCounts.IN_MITIGATION, 'text-amber-500'],
            ['completed', statusCounts.COMPLETED, 'text-blue-500'],
          ] as const).map(([key, value, color]) => (
            <div key={key}>
              <div className="text-[10px] font-semibold uppercase tracking-[0.16em] text-muted-foreground">
                {t(`darkWebMonitoring.summary.${key}`)}
              </div>
              <div className={`mt-1 text-xl font-semibold tabular-nums ${color}`}>
                {value.toLocaleString()}
              </div>
            </div>
          ))}
        </div>
      </section>

      <section className="overflow-hidden rounded-xl border border-border bg-card">
        <div className="flex items-center justify-between border-b border-border px-4 py-3">
          <h2 className="text-sm font-semibold">{t('darkWebMonitoring.surfaces.title')}</h2>
          <span className="text-[9px] font-semibold uppercase tracking-[0.16em] text-muted-foreground">
            {t('darkWebMonitoring.surfaces.exposed', { count: exposedSurfaceCount, total: SURFACES.length })}
          </span>
        </div>
        <button
          type="button"
          onClick={() => setSurface('domains')}
          className="flex w-full flex-wrap items-center gap-4 border-b border-border px-4 py-3 text-left transition-colors hover:bg-muted/20"
        >
          <span className="h-7 w-1 rounded-full bg-rose-500" />
          <span className="min-w-32 flex-1">
            <span className="block text-sm font-medium">{t('darkWebMonitoring.tabs.domains')}</span>
            <span className="text-[10px] text-muted-foreground">{t('darkWebMonitoring.surfaces.monitoredCount', { count: monitoredDomains.length })}</span>
            <span className="mt-1 flex flex-wrap gap-1.5">
              {monitoredDomains.map((domain) => (
                <span key={domain} className="rounded border border-border px-1.5 py-0.5 text-[10px] text-muted-foreground">{domain}</span>
              ))}
            </span>
          </span>
          <span className="text-right">
            <span className="block font-mono text-lg font-semibold tabular-nums">{surfaceCounts.domains.toLocaleString()}</span>
            <span className="text-[9px] uppercase tracking-wider text-muted-foreground">{t('darkWebMonitoring.summary.records')}</span>
          </span>
          <span className="text-xs font-medium text-primary">{t('darkWebMonitoring.surfaces.work')} <span aria-hidden="true">→</span></span>
        </button>
        <div className="flex flex-wrap items-center gap-2 px-4 py-3 text-[11px]">
          <span className="rounded-full border border-border px-2 py-1 text-muted-foreground">{t('darkWebMonitoring.surfaces.notMonitored')}</span>
          <span className="flex-1 text-muted-foreground">
            {t('darkWebMonitoring.tabs.emails')} · {t('darkWebMonitoring.tabs.ips')} · {t('darkWebMonitoring.tabs.usernames')} · {t('darkWebMonitoring.tabs.phones')}
          </span>
          <button type="button" onClick={() => setSurface('emails')} className="font-medium text-primary hover:underline">
            {t('darkWebMonitoring.surfaces.addAssets')} →
          </button>
        </div>
      </section>

      <section className="min-h-[380px] overflow-hidden rounded-xl border border-border bg-card">
        <div className="flex flex-wrap items-center justify-between gap-2 border-b border-border px-3">
          <div className="flex min-w-0 flex-wrap gap-1">
            {SURFACES.map((key) => (
              <button
                key={key}
                type="button"
                onClick={() => setSurface(key)}
                className={`relative flex items-center gap-1.5 px-2.5 py-3 text-xs transition-colors ${surface === key ? 'font-medium text-foreground' : 'text-muted-foreground hover:text-foreground'}`}
              >
                {t(`darkWebMonitoring.tabs.${key}`)}
                <span className={`rounded-full px-1.5 py-0.5 text-[9px] tabular-nums ${surface === key ? 'bg-primary/15 text-primary' : 'bg-muted text-muted-foreground'}`}>{surfaceCounts[key].toLocaleString()}</span>
                {surface === key && <span className="absolute inset-x-2 bottom-0 h-0.5 rounded-full bg-primary" />}
              </button>
            ))}
          </div>
          <Button size="sm" variant="outline" onClick={handleExport}>
            <Download size={13} className="mr-1.5" />
            {t('darkWebMonitoring.export')}
          </Button>
        </div>

        <div className="flex flex-wrap items-center gap-2 border-b border-border p-3">
          <label className="flex h-9 min-w-52 flex-1 items-center gap-2 rounded-lg border border-border bg-background px-2.5">
            <Search size={14} className="shrink-0 text-muted-foreground" />
            <input
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder={t('darkWebMonitoring.search')}
              className="w-full bg-transparent text-xs outline-none placeholder:text-muted-foreground"
            />
          </label>
          <div className="flex items-center gap-1 rounded-lg border border-border p-1">
            {STATUSES.map((key) => (
              <button
                key={key}
                type="button"
                onClick={() => setStatus(key)}
                className={`rounded-md px-2 py-1.5 text-[10px] transition-colors ${status === key ? 'bg-muted font-semibold text-foreground' : 'text-muted-foreground hover:text-foreground'}`}
              >
                {t(`darkWebMonitoring.statusFilter.${key === 'all' ? 'all' : STATUS_KEYS[key]}`)}
              </button>
            ))}
          </div>
          <TimeRangeFilter value={time} onChange={setTime} />
          <div className="ml-auto flex flex-wrap items-center gap-4 text-[10px] text-muted-foreground">
            <label className="flex cursor-pointer items-center gap-1.5">
              <input type="checkbox" checked={employeesOnly} onChange={(event) => setEmployeesOnly(event.target.checked)} />
              {t('darkWebMonitoring.employeeOnly')}
            </label>
            <label className="flex cursor-pointer items-center gap-1.5">
              <input type="checkbox" checked={reveal} onChange={(event) => setReveal(event.target.checked)} />
              <Eye size={12} />
              {t('darkWebMonitoring.reveal')}
            </label>
          </div>
        </div>

        <div
          ref={tableScrollRef}
          aria-label={t('darkWebMonitoring.table.results')}
          onScroll={handleTableScroll}
          className="max-h-[65vh] overflow-auto"
        >
          <table className="w-full min-w-[1000px] border-collapse text-left text-[11px]">
            <thead className="text-[9px] uppercase tracking-[0.14em] text-muted-foreground">
              <tr className="border-b border-border">
                <th colSpan={2} className="px-3 py-2 font-medium">{t('darkWebMonitoring.table.state')}</th>
                <th colSpan={7} className="border-l border-border px-3 py-2 font-medium">{t('darkWebMonitoring.table.whatLeaked')}</th>
                <th colSpan={4} className="border-l border-border px-3 py-2 font-medium">{t('darkWebMonitoring.table.whereFrom')}</th>
                <th colSpan={2} className="border-l border-border px-3 py-2 font-medium">{t('darkWebMonitoring.table.when')}</th>
              </tr>
              <tr className="border-b border-border/70">
                {(['status', 'who', 'domain', 'email', 'ipAddress', 'username', 'password', 'name', 'phone', 'database', 'leakName', 'foundIn', 'source', 'breached', 'ingested'] as const).map((key, index) => (
                  <th key={key} className={`whitespace-nowrap px-3 py-2 font-medium ${index === 2 || index === 9 || index === 13 ? 'border-l border-border' : ''}`}>
                    {t(`darkWebMonitoring.table.${key}`)}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-border/70">
              {visibleRecords.map((record) => (
                <tr
                  key={record.id}
                  tabIndex={0}
                  aria-label={t('darkWebMonitoring.details.openRecord', { value: record.organizationElement })}
                  onClick={() => setSelectedRecordId(record.id)}
                  onKeyDown={(event) => {
                    if (event.target !== event.currentTarget) return
                    if (event.key === 'Enter' || event.key === ' ') {
                      event.preventDefault()
                      setSelectedRecordId(record.id)
                    }
                  }}
                  className="cursor-pointer hover:bg-muted/20 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-inset focus-visible:ring-primary"
                >
                  <td className="whitespace-nowrap px-3 py-2.5">
                    <StatusPill
                      status={record.breachStatus}
                      label={t(`darkWebMonitoring.status.${STATUS_KEYS[record.breachStatus]}`)}
                      onChange={(nextStatus) => setRecordStatus(record.id, nextStatus)}
                    />
                  </td>
                  <td className="px-3 py-2.5">
                    <span className={`rounded-full px-2 py-1 text-[9px] ${record.employee ? 'bg-rose-500/10 text-rose-400' : 'bg-muted text-muted-foreground'}`}>
                      {record.employee ? t('darkWebMonitoring.employee') : t('darkWebMonitoring.other')}
                    </span>
                  </td>
                  <td className="whitespace-nowrap border-l border-border px-3 py-2.5 font-mono">{valueOrDash(record.domain)}</td>
                  <td className="whitespace-nowrap px-3 py-2.5 font-mono">{valueOrDash(record.email)}</td>
                  <td className="whitespace-nowrap px-3 py-2.5 font-mono">{valueOrDash(record.ipAddress)}</td>
                  <td className="whitespace-nowrap px-3 py-2.5 font-mono">{valueOrDash(record.username)}</td>
                  <td className="whitespace-nowrap px-3 py-2.5 font-mono text-muted-foreground">{reveal ? valueOrDash(getCredential(record)) : getCredential(record) ? MASKED : EMPTY_VALUE}</td>
                  <td className="whitespace-nowrap px-3 py-2.5">{valueOrDash(record.name)}</td>
                  <td className="whitespace-nowrap px-3 py-2.5 font-mono">{valueOrDash(record.phone)}</td>
                  <td className="whitespace-nowrap border-l border-border px-3 py-2.5 font-mono">{valueOrDash(record.databaseName)}</td>
                  <td className="whitespace-nowrap px-3 py-2.5">{valueOrDash(record.leakName)}</td>
                  <td className="whitespace-nowrap px-3 py-2.5">{valueOrDash(record.foundIn)}</td>
                  <td className="whitespace-nowrap px-3 py-2.5">{valueOrDash(record.source)}</td>
                  <td className="whitespace-nowrap border-l border-border px-3 py-2.5 font-mono">{formatDateShort(record.timestamp, i18n.language)}</td>
                  <td className="whitespace-nowrap px-3 py-2.5 font-mono">{formatDateShort(record.createdAt, i18n.language)}</td>
                </tr>
              ))}
              {visibleRecords.length === 0 && (
                <tr><td colSpan={15} className="px-4 py-12 text-center text-sm text-muted-foreground">{t('darkWebMonitoring.empty')}</td></tr>
              )}
            </tbody>
          </table>
          {loadingMore && (
            <div className="flex items-center justify-center gap-2 py-4 text-xs text-muted-foreground" role="status">
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
              {t('darkWebMonitoring.loadingMore')}
            </div>
          )}
          {!hasMoreRecords && filtered.length > 0 && (
            <div className="py-4 text-center text-[11px] text-muted-foreground">
              {t('darkWebMonitoring.endOfResults', { count: filtered.length.toLocaleString() })}
            </div>
          )}
        </div>
      </section>
    </div>
  )
}
