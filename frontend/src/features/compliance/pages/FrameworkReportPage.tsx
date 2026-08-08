import { useCallback, useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, ArrowLeft, Download, Loader2, Play } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { complianceService } from '../services/compliance-http.service'
import type { ControlRow, Report } from '../types/compliance.types'
import { ReportView } from '../components/ReportView'
import { ControlDetailDrawer } from '../components/ControlDetailDrawer'
import { ReportPreview } from '../components/ReportPreview'
import { ScoreHistory } from '../components/ScoreHistory'

export function FrameworkReportPage() {
  const { key = '' } = useParams<{ key: string }>()
  const navigate = useNavigate()
  const { t } = useTranslation()
  const [report, setReport] = useState<Report | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [openRow, setOpenRow] = useState<ControlRow | null>(null)
  const [running, setRunning] = useState(false)
  const [preview, setPreview] = useState(false)

  const load = useCallback(() => {
    if (!key) return
    setLoading(true)
    setError(false)
    // Reads the standing report; it does not re-run the framework. Running is
    // an explicit action, because a run rewrites what a person may have edited.
    complianceService
      .getReport(key)
      .then(setReport)
      .catch(() => setError(true))
      .finally(() => setLoading(false))
  }, [key])
  useEffect(() => {
    load()
  }, [load])

  const back = () => navigate('/compliance')

  /** An edit returns the whole recomputed report, so nothing has to be refetched. */
  const applyUpdated = (fresh: Report) => {
    setReport(fresh)
    if (openRow) {
      const stillHere = fresh.controls?.find((c) => c.controlId === openRow.controlId)
      if (stillHere) setOpenRow(stillHere)
    }
  }

  const run = async () => {
    setRunning(true)
    try {
      applyUpdated(await complianceService.evaluate(key))
    } catch {
      toast.error(t('compliance.runError', { defaultValue: 'Evaluation failed' }))
    } finally {
      setRunning(false)
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
        <div className="flex shrink-0 items-center gap-2">
          <Button size="sm" variant="outline" onClick={() => void run()} disabled={running}>
            {running ? <Loader2 size={14} className="mr-1.5 animate-spin" /> : <Play size={14} className="mr-1.5" />}
            {t('compliance.run', { defaultValue: 'Run evaluation' })}
          </Button>
          <Button size="sm" onClick={() => setPreview(true)} disabled={!report}>
            <Download size={14} className="mr-1.5" />
            {t('compliance.doc.download')}
          </Button>
        </div>
      </header>

      <div className="mt-4 min-h-0 flex-1 overflow-y-auto">
        {loading ? (
          <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" /> {t('compliance.evaluating')}
          </div>
        ) : !report ? (
          <div className="flex flex-col items-center justify-center gap-3 px-6 py-16 text-sm text-muted-foreground">
            {error && <AlertTriangle size={16} className="text-amber-500" />}
            {t('compliance.noReportYet', { defaultValue: 'This framework has not been evaluated yet.' })}
            <Button size="sm" onClick={() => void run()} disabled={running}>
              {running ? <Loader2 size={14} className="mr-1.5 animate-spin" /> : <Play size={14} className="mr-1.5" />}
              {t('compliance.run', { defaultValue: 'Run evaluation' })}
            </Button>
          </div>
        ) : (
          <>
            {/* Above the findings: how the position got here frames what
                follows, and a chart at the foot of a 155-row list is never
                reached. */}
            <ScoreHistory frameworkKey={key} />
            <div className="mt-5">
              <ReportView report={report} onControlClick={setOpenRow} onChanged={applyUpdated} />
            </div>
          </>
        )}
      </div>

      {preview && report && (
        <ReportPreview
          frameworkKey={key}
          frameworkName={report.frameworkName}
          onClose={() => setPreview(false)}
        />
      )}

      {openRow && (
        <ControlDetailDrawer
          frameworkKey={key}
          row={openRow}
          onClose={() => setOpenRow(null)}
          onStatusChanged={applyUpdated}
        />
      )}

    </div>
  )
}
