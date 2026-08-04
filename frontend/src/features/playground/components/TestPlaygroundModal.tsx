import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, Info, Loader2, PlayCircle, X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { PlaygroundHttpError, playgroundHttpService } from '../services/playground-http.service'
import type { PlaygroundMode, TestFilterResponse } from '../types'
import { ParsedEventView } from './ParsedEventView'
import { AlertsListView } from './AlertsListView'

type ErrorKind = 'busy' | 'timeout' | 'validation' | 'server' | null

// Mirrors the backend's raw-log body limit — shown client-side to preempt a
// 413 rather than let the user discover it only after Run fails.
const MAX_BYTES = 1024 * 1024

interface Props {
  mode: PlaygroundMode
  /** i18n key for the modal header, e.g. `playground.titleFilter` or `playground.titleRule`. */
  titleKey: string
  /** Entry A: active dataTypes from the catalog. Entry B: `model.dataTypes` of the draft. */
  dataTypeOptions: string[]
  /** Present only for Entry B (unsaved draft content); omitted tests the live pipeline. */
  draftContent?: string
  onClose: () => void
}

export function TestPlaygroundModal({ mode, titleKey, dataTypeOptions, draftContent, onClose }: Props) {
  const { t } = useTranslation()
  const [selectedDataType, setSelectedDataType] = useState(dataTypeOptions[0] ?? '')
  const [rawLog, setRawLog] = useState('')
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<TestFilterResponse | null>(null)
  const [errorKind, setErrorKind] = useState<ErrorKind>(null)
  const [elapsedMs, setElapsedMs] = useState<number | null>(null)

  const noDataTypes = dataTypeOptions.length === 0
  // Entry B (draft) usually has exactly one dataType — showing an interactive
  // dropdown implies a choice that doesn't exist, so lock it instead.
  const singleDataType = dataTypeOptions.length === 1
  const byteSize = new TextEncoder().encode(rawLog).length
  const overLimit = byteSize > MAX_BYTES
  const nearLimit = !overLimit && byteSize > MAX_BYTES * 0.9
  const runDisabled = loading || !rawLog.trim() || !selectedDataType || noDataTypes || overLimit

  const run = async () => {
    if (runDisabled) return
    setLoading(true)
    setErrorKind(null)
    setResult(null)
    setElapsedMs(null)
    const start = performance.now()
    try {
      const log = { dataType: selectedDataType, raw: rawLog }
      const response =
        mode === 'rule'
          ? await playgroundHttpService.testRule({ log, rule: draftContent !== undefined ? { content: draftContent } : undefined })
          : await playgroundHttpService.testFilter({ log, filter: draftContent !== undefined ? { content: draftContent } : undefined })
      setElapsedMs(performance.now() - start)
      if (response.timedOut) {
        setErrorKind('timeout')
      } else {
        setResult(response)
      }
    } catch (err) {
      if (err instanceof PlaygroundHttpError && err.status === 503) setErrorKind('busy')
      else if (err instanceof PlaygroundHttpError && err.status === 400) setErrorKind('validation')
      // 500/502 map here explicitly; everything else (413, network, etc.)
      // falls back to the same generic server copy.
      else setErrorKind('server')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm"
      onClick={(e) => {
        // This modal may be nested inside another overlay (e.g. FilterFormDrawer);
        // stop the click here so it doesn't also close the parent.
        e.stopPropagation()
        onClose()
      }}
    >
      <div
        className="flex h-[90vh] w-[95vw] max-w-[1600px] flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-center justify-between gap-4 border-b border-border px-6 py-4">
          <h2 className="flex items-center gap-2 text-lg font-semibold">
            <PlayCircle size={16} className="text-primary" /> {t(titleKey)}
          </h2>
          <button
            type="button"
            onClick={onClose}
            className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <X size={16} />
          </button>
        </header>

        <div className="grid min-h-0 flex-1 grid-cols-1 gap-6 overflow-y-auto px-6 py-5 md:grid-cols-2 md:overflow-hidden">
          {/* Left column: test inputs */}
          <div className="flex min-h-0 flex-col gap-4 md:overflow-y-auto md:pr-1">
            {noDataTypes && (
              <div className="flex items-center gap-2 rounded-md bg-amber-500/10 px-3 py-2 text-xs text-amber-600 dark:text-amber-300">
                <Info size={13} /> {t('playground.noDataTypes')}
              </div>
            )}

            <div className="space-y-1.5">
              <label className="block text-xs font-medium text-foreground/80">{t('playground.dataTypeLabel')}</label>
              <select
                value={selectedDataType}
                onChange={(e) => setSelectedDataType(e.target.value)}
                disabled={noDataTypes || singleDataType}
                title={singleDataType ? t('playground.dataTypeSingleHint') : undefined}
                className="h-9 w-full rounded-md border border-input bg-background px-2.5 text-xs text-foreground focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:opacity-50"
              >
                {noDataTypes ? (
                  <option value="">{t('playground.dataTypePlaceholder')}</option>
                ) : (
                  dataTypeOptions.map((dt) => (
                    <option key={dt} value={dt}>
                      {dt}
                    </option>
                  ))
                )}
              </select>
            </div>

            <div className="space-y-1.5">
              <div className="flex items-center justify-between">
                <label className="block text-xs font-medium text-foreground/80">{t('playground.rawLogLabel')}</label>
                <span className={cn('text-[11px]', overLimit ? 'text-red-500' : nearLimit ? 'text-amber-500' : 'text-muted-foreground')}>
                  {t('playground.rawLogCounter', { chars: rawLog.length, kb: (byteSize / 1024).toFixed(1) })}
                </span>
              </div>
              <textarea
                value={rawLog}
                onChange={(e) => setRawLog(e.target.value)}
                rows={12}
                spellCheck={false}
                autoCorrect="off"
                autoCapitalize="off"
                placeholder={t('playground.rawLogPlaceholder')}
                className="w-full resize-none rounded-md border border-input bg-background/40 p-2 font-mono text-xs focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              />
              {overLimit && <p className="text-[11px] text-red-500">{t('playground.rawLogTooLarge')}</p>}
            </div>

            <div className="flex items-center gap-2">
              <Button size="sm" onClick={() => void run()} disabled={runDisabled}>
                {loading ? <Loader2 size={13} className="mr-1.5 animate-spin" /> : <PlayCircle size={13} className="mr-1.5" />}
                {loading ? t('playground.run.running') : t('playground.run.label')}
              </Button>
              {loading && <span className="text-[11px] text-muted-foreground">{t('playground.run.waitHint')}</span>}
            </div>

            {errorKind && (
              <div className="flex items-start gap-2 rounded-md bg-red-500/10 px-3 py-2 text-xs text-red-600 dark:text-red-300">
                <AlertTriangle size={13} className="mt-0.5 shrink-0" /> {t(`playground.errors.${errorKind}`)}
              </div>
            )}

            {mode === 'rule' && (
              <>
                <div className="flex items-start gap-2 rounded-md bg-sky-500/10 px-3 py-2 text-xs text-sky-700 dark:text-sky-300">
                  <Info size={13} className="mt-0.5 shrink-0" /> {t('playground.notices.geo')}
                </div>
                <div className="flex items-start gap-2 rounded-md bg-sky-500/10 px-3 py-2 text-xs text-sky-700 dark:text-sky-300">
                  <Info size={13} className="mt-0.5 shrink-0" /> {t('playground.notices.correlation')}
                </div>
              </>
            )}
          </div>

          {/* Right column: result */}
          <div className="flex min-h-0 flex-col gap-3 md:overflow-y-auto md:pl-1">
            <div className={cn('flex min-h-0 flex-col gap-1.5', mode === 'rule' ? 'flex-[3]' : 'flex-1')}>
              <label className="block text-xs font-medium text-foreground/80">{t('playground.result.label')}</label>
              <div className="min-h-0 flex-1">
                <ParsedEventView event={result?.event} loading={loading} elapsedMs={elapsedMs ?? undefined} />
              </div>
            </div>
            {mode === 'rule' && (
              <div className="flex min-h-0 flex-[2] flex-col gap-1.5">
                <label className="block text-xs font-medium text-foreground/80">{t('playground.alerts.label')}</label>
                <div className="min-h-0 flex-1">
                  <AlertsListView alerts={result?.alerts} loading={loading} />
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
