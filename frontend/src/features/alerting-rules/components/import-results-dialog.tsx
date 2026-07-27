import type { TFunction } from 'i18next'
import { CheckCircle2, Upload, X, XCircle } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import type { ImportRulesResponse } from '../services/alerting-rules-http.service'

export function ImportResultsDialog({ res, onClose, t }: { res: ImportRulesResponse; onClose: () => void; t: TFunction }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm" onClick={onClose}>
      <div className="flex max-h-[80vh] w-full max-w-[560px] flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl" onClick={(e) => e.stopPropagation()}>
        <header className="flex items-center justify-between border-b border-border px-5 py-3.5">
          <h2 className="flex items-center gap-2 text-base font-semibold">
            <Upload size={16} /> {t('alertingRules.import.resultTitle')}
          </h2>
          <button onClick={onClose} className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground">
            <X size={16} />
          </button>
        </header>

        <div className="flex shrink-0 items-center gap-4 border-b border-border bg-muted/20 px-5 py-2.5 text-xs">
          <span className="inline-flex items-center gap-1.5 text-emerald-600 dark:text-emerald-400">
            <CheckCircle2 size={14} /> {t('alertingRules.import.approved')}: <b>{res.approved}</b>
          </span>
          <span className="inline-flex items-center gap-1.5 text-red-600 dark:text-red-400">
            <XCircle size={14} /> {t('alertingRules.import.rejected')}: <b>{res.rejected}</b>
          </span>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto p-3">
          <div className="space-y-1.5">
            {res.results.map((r, i) => (
              <div key={i} className="flex items-start gap-2 rounded-md border border-border px-3 py-2 text-xs">
                {r.approved ? (
                  <CheckCircle2 size={15} className="mt-0.5 shrink-0 text-emerald-500" />
                ) : (
                  <XCircle size={15} className="mt-0.5 shrink-0 text-red-500" />
                )}
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center gap-x-2">
                    <span className="font-mono text-[11px] text-muted-foreground">{r.filename}</span>
                    {r.name && <span className="font-medium">{r.name}</span>}
                  </div>
                  {!r.approved && r.error && <p className="mt-0.5 text-[11px] text-red-600 dark:text-red-400">{r.error}</p>}
                </div>
              </div>
            ))}
          </div>
        </div>

        <footer className="flex justify-end border-t border-border px-5 py-3">
          <Button size="sm" onClick={onClose}>{t('alertingRules.import.done')}</Button>
        </footer>
      </div>
    </div>
  )
}
