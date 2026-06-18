import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, FileBarChart, Loader2, RefreshCw, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { useDateFormat } from '@/shared/lib/datetime'
import { complianceService } from '../services/compliance-http.service'
import type { ReportSnapshot, ReportSnapshotMeta } from '../types/compliance.types'
import { scoreTone } from './ReportView'
import { ReportDocument } from './ReportDocument'

export function ReportsTab() {
  const { t } = useTranslation()
  const df = useDateFormat()
  const [snaps, setSnaps] = useState<ReportSnapshotMeta[]>([])
  const [lockedKeys, setLockedKeys] = useState<Set<string>>(new Set())
  const [fw, setFw] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [open, setOpen] = useState<ReportSnapshot | null>(null)
  const [opening, setOpening] = useState(false)
  const [confirmId, setConfirmId] = useState<string | null>(null)
  const [deleting, setDeleting] = useState(false)

  const load = useCallback(() => {
    setLoading(true)
    setError(false)
    Promise.all([complianceService.listSnapshots(undefined, 200), complianceService.listFrameworks()])
      .then(([s, fws]) => {
        setSnaps(s ?? [])
        setLockedKeys(new Set((fws ?? []).filter((f) => f.locked).map((f) => f.key)))
      })
      .catch(() => setError(true))
      .finally(() => setLoading(false))
  }, [])
  useEffect(() => {
    load()
  }, [load])

  // Hide snapshots of frameworks the current edition can't use (e.g. left over from
  // a prior enterprise license).
  const visible = useMemo(() => snaps.filter((s) => !lockedKeys.has(s.frameworkKey)), [snaps, lockedKeys])

  const frameworkOptions = useMemo(() => {
    const m = new Map<string, string>()
    for (const s of visible) if (!m.has(s.frameworkKey)) m.set(s.frameworkKey, s.frameworkName)
    return [...m.entries()]
  }, [visible])

  const shown = useMemo(() => (fw ? visible.filter((s) => s.frameworkKey === fw) : visible), [visible, fw])

  const openSnapshot = async (id: string) => {
    setOpening(true)
    try {
      setOpen(await complianceService.getSnapshot(id))
    } catch {
      setError(true)
    } finally {
      setOpening(false)
    }
  }

  const remove = async (id: string) => {
    setDeleting(true)
    try {
      await complianceService.deleteSnapshot(id)
      setSnaps((l) => l.filter((x) => x.id !== id))
      setConfirmId(null)
      toast.success(t('compliance.reports.toast.deleted'))
    } catch {
      toast.error(t('compliance.reports.toast.deleteError'))
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="mb-3 flex shrink-0 items-center gap-2">
        <select
          value={fw}
          onChange={(e) => setFw(e.target.value)}
          className="h-9 rounded-md border border-input bg-background px-2.5 text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
        >
          <option value="">{t('compliance.reports.allFrameworks')}</option>
          {frameworkOptions.map(([key, name]) => (
            <option key={key} value={key}>{name}</option>
          ))}
        </select>
        <Button variant="outline" size="sm" onClick={load} disabled={loading} title={t('compliance.refresh')}>
          <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
        </Button>
        {opening && <Loader2 size={14} className="animate-spin text-muted-foreground" />}
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto rounded-xl border border-border bg-card">
        {loading && snaps.length === 0 ? (
          <Center><Loader2 className="h-4 w-4 animate-spin" /> {t('compliance.reports.loading')}</Center>
        ) : error ? (
          <Center><AlertTriangle size={16} className="text-amber-500" /> {t('compliance.reports.loadError')}<Button variant="outline" size="sm" className="ml-2" onClick={load}>{t('compliance.retry')}</Button></Center>
        ) : shown.length === 0 ? (
          <div className="px-6 py-16 text-center text-sm text-muted-foreground">{t('compliance.reports.empty')}</div>
        ) : (
          shown.map((s) => (
            <div
              key={s.id}
              className="flex w-full items-center gap-3 border-b border-border px-4 py-3 transition-colors last:border-0 hover:bg-muted/40"
            >
              <button onClick={() => void openSnapshot(s.id)} className="flex min-w-0 flex-1 items-center gap-3 text-left">
                <FileBarChart size={15} className="shrink-0 text-muted-foreground" />
                <div className="min-w-0 flex-1">
                  <div className="truncate text-[13px] font-medium">{s.frameworkName}</div>
                  <div className="text-[11px] text-muted-foreground">{df.formatDateTime(s['@timestamp'])}</div>
                </div>
              </button>
              <div className={cn('text-lg font-bold tabular-nums', scoreTone(s.score))}>{s.score}%</div>
              {confirmId === s.id ? (
                <div className="flex items-center gap-1">
                  <Button size="sm" variant="destructive" className="h-7" disabled={deleting} onClick={() => void remove(s.id)}>
                    {deleting ? <Loader2 size={12} className="mr-1 animate-spin" /> : <Trash2 size={12} className="mr-1" />}
                    {t('compliance.reports.confirmDelete')}
                  </Button>
                  <Button size="sm" variant="outline" className="h-7" disabled={deleting} onClick={() => setConfirmId(null)}>
                    {t('compliance.reports.cancel')}
                  </Button>
                </div>
              ) : (
                <button
                  onClick={() => setConfirmId(s.id)}
                  title={t('compliance.reports.delete')}
                  className="rounded-md p-1.5 text-muted-foreground hover:bg-red-500/10 hover:text-red-500"
                >
                  <Trash2 size={14} />
                </button>
              )}
            </div>
          ))
        )}
      </div>

      {open && <ReportDocument report={open.report} onClose={() => setOpen(null)} onDownloadPdf={() => complianceService.downloadSnapshotPdf(open.id)} />}
    </div>
  )
}

function Center({ children }: { children: React.ReactNode }) {
  return <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">{children}</div>
}
