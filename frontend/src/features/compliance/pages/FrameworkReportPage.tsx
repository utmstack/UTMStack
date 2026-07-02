import { useCallback, useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, ArrowLeft, Download, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { complianceService } from '../services/compliance-http.service'
import type { Report, ReportControlRow } from '../types/compliance.types'
import { ReportView } from '../components/ReportView'
import { ControlDetailDrawer } from '../components/ControlDetailDrawer'

export function FrameworkReportPage() {
  const { key = '' } = useParams<{ key: string }>()
  const navigate = useNavigate()
  const { t } = useTranslation()
  const [report, setReport] = useState<Report | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [downloading, setDownloading] = useState(false)
  const [openRow, setOpenRow] = useState<ReportControlRow | null>(null)

  const load = useCallback(() => {
    if (!key) return
    setLoading(true)
    setError(false)
    complianceService
      .liveReport(key)
      .then(setReport)
      .catch(() => setError(true))
      .finally(() => setLoading(false))
  }, [key])
  useEffect(() => {
    load()
  }, [load])

  const back = () => navigate('/compliance')

  const reloadAndResyncDrawer = async () => {
    const fresh = await complianceService.liveReport(key).catch(() => null)
    if (!fresh) return
    setReport(fresh)
    if (openRow) {
      const stillHere = fresh.sections?.flatMap((s) => s.controls ?? []).find((c) => c.controlId === openRow.controlId)
      if (stillHere) setOpenRow(stillHere)
    }
  }

  const download = async () => {
    if (downloading) return
    setDownloading(true)
    try {
      const blob = await complianceService.downloadFrameworkPdf(key)
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${(report?.frameworkName || key).replace(/[^\w -]/g, '').trim() || 'Compliance_Report'}.pdf`
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

  return (
    <div className="mx-auto flex h-full min-h-0 w-full max-w-[1100px] flex-col px-6 pb-6 pt-3">
      <header className="flex shrink-0 items-center justify-between gap-3">
        <Button variant="ghost" size="sm" onClick={back}>
          <ArrowLeft size={14} className="mr-1.5" /> {t('compliance.backToFrameworks')}
        </Button>
        <div className="min-w-0 flex-1 text-center">
          <h1 className="truncate text-sm font-semibold">{report?.frameworkName || key}</h1>
        </div>
        <Button size="sm" onClick={() => void download()} disabled={downloading || !report}>
          {downloading ? <Loader2 size={14} className="mr-1.5 animate-spin" /> : <Download size={14} className="mr-1.5" />}
          {t('compliance.doc.download')}
        </Button>
      </header>

      <div className="mt-4 min-h-0 flex-1 overflow-y-auto">
        {loading ? (
          <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" /> {t('compliance.evaluating')}
          </div>
        ) : error || !report ? (
          <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">
            <AlertTriangle size={16} className="text-amber-500" /> {t('compliance.reportError')}
            <Button variant="outline" size="sm" className="ml-2" onClick={load}>{t('compliance.retry')}</Button>
          </div>
        ) : (
          <ReportView report={report} onControlClick={setOpenRow} onChanged={reloadAndResyncDrawer} />
        )}
      </div>

      {openRow && (
        <ControlDetailDrawer
          frameworkKey={key}
          row={openRow}
          onClose={() => setOpenRow(null)}
          onStatusChanged={reloadAndResyncDrawer}
        />
      )}
    </div>
  )
}
