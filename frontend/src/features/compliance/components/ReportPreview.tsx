import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, Download, Loader2, X } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { complianceService } from '../services/compliance-http.service'

/**
 * Shows the report before it is saved.
 *
 * It previews the actual PDF rather than an HTML lookalike. A second renderer
 * of the same document is a second thing to keep in step, and the one that
 * lived here had already drifted — the preview showed one document and the
 * downloaded file held another. Here the bytes on screen are the bytes saved.
 */
export function ReportPreview({
  frameworkKey,
  frameworkName,
  onClose,
}: {
  frameworkKey: string
  frameworkName: string
  onClose: () => void
}) {
  const { t } = useTranslation()
  const [url, setUrl] = useState<string | null>(null)
  const [blob, setBlob] = useState<Blob | null>(null)
  const [error, setError] = useState(false)

  useEffect(() => {
    let objectUrl: string | null = null
    let cancelled = false
    complianceService
      .downloadReportPdf(frameworkKey)
      .then((b) => {
        if (cancelled) return
        objectUrl = URL.createObjectURL(b)
        setBlob(b)
        setUrl(objectUrl)
      })
      .catch(() => !cancelled && setError(true))
    return () => {
      cancelled = true
      // Released on close: an object URL holds the whole file in memory until
      // it is revoked, and a report runs to several hundred kilobytes.
      if (objectUrl) URL.revokeObjectURL(objectUrl)
    }
  }, [frameworkKey])

  const save = () => {
    if (!blob) return
    const a = document.createElement('a')
    const href = URL.createObjectURL(blob)
    a.href = href
    a.download = `${frameworkName.replace(/[^A-Za-z0-9]+/g, '_')}.pdf`
    a.click()
    URL.revokeObjectURL(href)
  }

  return (
    <div className="fixed inset-0 z-50 flex flex-col bg-black/60 backdrop-blur-sm" onClick={onClose}>
      <div
        className="mx-auto my-4 flex h-full w-full max-w-[900px] flex-col overflow-hidden rounded-xl border border-border bg-card"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex shrink-0 items-center justify-between gap-3 border-b border-border px-5 py-3">
          <h2 className="truncate text-sm font-semibold">{frameworkName}</h2>
          <div className="flex items-center gap-2">
            <Button size="sm" onClick={save} disabled={!blob}>
              <Download size={14} className="mr-1.5" />
              {t('compliance.doc.download')}
            </Button>
            <button
              onClick={onClose}
              aria-label={t('compliance.schedule.cancel')}
              className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <X size={16} />
            </button>
          </div>
        </header>

        <div className="min-h-0 flex-1 bg-muted/40">
          {error ? (
            <div className="flex h-full items-center justify-center gap-2 text-sm text-muted-foreground">
              <AlertTriangle size={16} className="text-amber-500" />
              {t('compliance.doc.error', { defaultValue: 'Could not generate the PDF' })}
            </div>
          ) : !url ? (
            <div className="flex h-full items-center justify-center gap-2 text-sm text-muted-foreground">
              <Loader2 className="h-4 w-4 animate-spin" />
              {t('compliance.doc.building', { defaultValue: 'Building the document…' })}
            </div>
          ) : (
            <iframe src={url} title={frameworkName} className="h-full w-full border-0" />
          )}
        </div>
      </div>
    </div>
  )
}
