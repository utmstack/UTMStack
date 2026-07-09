import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, FileText, Loader2, Lock, X } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { InfiniteScrollSentinel } from '@/shared/components/ui/infinite-scroll'
import { soarTemplatesService } from '../services/soar-templates.service'
import type { ActionTemplate } from '../types/soar.types'

const PAGE_SIZE = 20

export function TemplatePicker({
  onPick,
  onScratch,
  onClose,
}: {
  onPick: (t: ActionTemplate) => void
  onScratch: () => void
  onClose: () => void
}) {
  const { t } = useTranslation()
  const [items, setItems] = useState<ActionTemplate[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  const loadPage = useCallback(async (p: number) => {
    setLoading(true)
    setError(false)
    try {
      const res = await soarTemplatesService.list({ page: p, size: PAGE_SIZE })
      setTotal(res.total)
      setItems((prev) => (p === 0 ? res.data : [...prev, ...res.data]))
    } catch {
      setError(true)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadPage(page)
  }, [page, loadPage])

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div className="flex max-h-[80vh] w-full max-w-[640px] flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl" onClick={(e) => e.stopPropagation()}>
        <header className="flex items-start justify-between gap-4 border-b border-border px-5 py-3">
          <div className="min-w-0">
            <h2 className="truncate text-base font-semibold">{t('soar.templates.title')}</h2>
            <p className="mt-0.5 text-[11px] text-muted-foreground">{t('soar.templates.subtitle')}</p>
          </div>
          <button onClick={onClose} className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground">
            <X size={15} />
          </button>
        </header>

        <div className="min-h-0 flex-1 overflow-y-auto">
          {loading && items.length === 0 ? (
            <div className="flex items-center justify-center gap-2 px-6 py-16 text-xs text-muted-foreground">
              <Loader2 className="h-4 w-4 animate-spin" /> {t('soar.templates.loading')}
            </div>
          ) : error && items.length === 0 ? (
            <div className="flex items-center justify-center gap-2 px-6 py-16 text-xs text-muted-foreground">
              <AlertTriangle size={14} className="text-amber-500" /> {t('soar.templates.loadError')}
              <Button variant="outline" size="sm" className="ml-2" onClick={() => loadPage(0)}>{t('soar.retry')}</Button>
            </div>
          ) : items.length === 0 ? (
            <div className="px-6 py-16 text-center text-xs text-muted-foreground">{t('soar.templates.empty')}</div>
          ) : (
            <div className="divide-y divide-border">
              {items.map((tpl) => (
                <button
                  key={tpl.id}
                  type="button"
                  onClick={() => onPick(tpl)}
                  className="flex w-full items-start gap-3 px-5 py-3 text-left transition-colors hover:bg-muted/50"
                >
                  <FileText size={14} className="mt-0.5 shrink-0 text-muted-foreground" />
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                      <span className="truncate text-sm font-medium">{tpl.title}</span>
                      {tpl.systemOwner && (
                        <span className="inline-flex items-center gap-1 rounded bg-violet-500/15 px-1.5 py-0.5 text-[10px] font-medium text-violet-500">
                          <Lock size={9} /> {t('soar.system')}
                        </span>
                      )}
                    </div>
                    {tpl.description && (
                      <p className="mt-0.5 line-clamp-2 text-[11px] text-muted-foreground">{tpl.description}</p>
                    )}
                  </div>
                </button>
              ))}
              <InfiniteScrollSentinel
                onReach={() => setPage((p) => p + 1)}
                hasMore={items.length < total}
                loading={loading}
                endLabel={t('common.allLoaded', { count: total })}
              />
            </div>
          )}
        </div>

        <footer className="flex items-center justify-end gap-2 border-t border-border px-5 py-3">
          <Button variant="ghost" size="sm" onClick={onClose}>{t('soar.templates.cancel')}</Button>
          <Button size="sm" onClick={onScratch}>{t('soar.templates.scratch')}</Button>
        </footer>
      </div>
    </div>
  )
}
