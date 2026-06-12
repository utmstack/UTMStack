import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Loader2, Plus, Search, X } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { Pagination } from '@/shared/components/ui/pagination'
import { cn } from '@/shared/lib/utils'
import { useVisualizations } from '@/features/dashboard/hooks/useVisualizations'
import { DEFAULT_PAGE_SIZE } from '@/features/dashboard/constants'
import type { Visualization } from '@/features/dashboard/types'

export function AddVisualizationDrawer({
  open,
  excludedIds,
  busy,
  onClose,
  onAdd,
}: {
  open: boolean
  excludedIds: Set<number>
  busy?: boolean
  onClose: () => void
  onAdd: (visualizations: Visualization[]) => void | Promise<void>
}) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(0)
  const [size, setSize] = useState(DEFAULT_PAGE_SIZE)
  const [selected, setSelected] = useState<Map<number, Visualization>>(new Map())

  useEffect(() => {
    if (!open) setSelected(new Map())
  }, [open])

  const query = useVisualizations({ name: search || undefined, page, size })

  if (!open) return null

  const items = query.data?.data ?? []
  const total = query.data?.total ?? 0
  const selectedCount = selected.size

  const toggle = (viz: Visualization) => {
    if (excludedIds.has(viz.id)) return
    setSelected((prev) => {
      const next = new Map(prev)
      if (next.has(viz.id)) {
        next.delete(viz.id)
      } else {
        next.set(viz.id, viz)
      }
      return next
    })
  }

  const submit = async () => {
    if (selectedCount === 0 || busy) return
    await onAdd(Array.from(selected.values()))
  }

  return (
    <div
      className="fixed inset-0 z-40 flex justify-end bg-black/40 backdrop-blur-sm"
      onClick={() => !busy && onClose()}
    >
      <div
        className="flex h-full w-full max-w-md flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-start justify-between gap-4 border-b border-border px-5 py-4">
          <div>
            <h2 className="text-base font-semibold">{t('dashboards.addWidget.title')}</h2>
            <p className="mt-1 text-xs text-muted-foreground">
              {t('dashboards.addWidget.subtitle')}
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            disabled={busy}
            className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50"
            aria-label={t('common.actions.cancel') ?? 'Cancel'}
          >
            <X size={16} />
          </button>
        </header>

        <div className="flex items-center gap-2 border-b border-border px-5 py-3">
          <div className="relative flex-1">
            <Search
              size={14}
              className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground"
            />
            <Input
              value={search}
              onChange={(e) => {
                setSearch(e.target.value)
                setPage(0)
              }}
              placeholder={t('dashboards.addWidget.searchPlaceholder') ?? ''}
              className="h-9 pl-9"
            />
          </div>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => navigate('/dashboards/visualizations/new')}
          >
            <Plus size={14} className="mr-1" />
            {t('dashboards.addWidget.newWidget')}
          </Button>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto px-2 py-2">
          {query.isLoading && (
            <div className="flex items-center justify-center gap-2 px-3 py-8 text-xs text-muted-foreground">
              <Loader2 size={14} className="animate-spin" />
              {t('dashboards.addWidget.loading')}
            </div>
          )}
          {!query.isLoading && items.length === 0 && (
            <div className="px-3 py-8 text-center text-xs text-muted-foreground">
              {t('dashboards.addWidget.empty')}
            </div>
          )}
          <ul className="space-y-1">
            {items.map((viz) => {
              const already = excludedIds.has(viz.id)
              const checked = selected.has(viz.id)
              return (
                <li key={viz.id}>
                  <button
                    type="button"
                    disabled={already}
                    onClick={() => toggle(viz)}
                    className={cn(
                      'flex w-full items-start gap-3 rounded-md border px-3 py-2 text-left transition-colors disabled:cursor-not-allowed disabled:opacity-50',
                      checked
                        ? 'border-primary bg-primary/5'
                        : 'border-transparent hover:border-border hover:bg-muted'
                    )}
                  >
                    <span
                      className={cn(
                        'mt-0.5 flex h-4 w-4 shrink-0 items-center justify-center rounded border',
                        checked
                          ? 'border-primary bg-primary text-primary-foreground'
                          : 'border-input bg-background'
                      )}
                    >
                      {checked && (
                        <svg
                          width="10"
                          height="10"
                          viewBox="0 0 12 12"
                          fill="none"
                          xmlns="http://www.w3.org/2000/svg"
                        >
                          <path
                            d="M2 6.5L4.5 9L10 3"
                            stroke="currentColor"
                            strokeWidth="2"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                          />
                        </svg>
                      )}
                    </span>
                    <span className="flex min-w-0 flex-1 flex-col gap-0.5">
                      <span className="truncate text-sm font-medium">{viz.name}</span>
                      {viz.description && (
                        <span className="truncate text-xs text-muted-foreground">
                          {viz.description}
                        </span>
                      )}
                      {already && (
                        <span className="text-[10px] uppercase tracking-wide text-muted-foreground/70">
                          {t('dashboards.addWidget.alreadyAdded')}
                        </span>
                      )}
                    </span>
                  </button>
                </li>
              )
            })}
          </ul>
        </div>

        <div className="border-t border-border px-5 py-3">
          <Pagination
            page={page}
            pageSize={size}
            total={total}
            loading={query.isLoading}
            onPageChange={setPage}
            onPageSizeChange={(s) => {
              setSize(s)
              setPage(0)
            }}
            align="right"
          />
        </div>

        <footer className="flex items-center justify-between gap-2 border-t border-border bg-muted/30 px-5 py-3">
          <span className="text-xs text-muted-foreground">
            {selectedCount > 0
              ? t('dashboards.addWidget.selectedCount', { count: selectedCount })
              : t('dashboards.addWidget.noneSelected')}
          </span>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" onClick={onClose} disabled={busy}>
              {t('common.actions.cancel')}
            </Button>
            <Button
              size="sm"
              onClick={() => void submit()}
              disabled={selectedCount === 0 || busy}
            >
              {busy && <Loader2 size={14} className="mr-1 animate-spin" />}
              {t('dashboards.addWidget.addSelected', { count: selectedCount })}
            </Button>
          </div>
        </footer>
      </div>
    </div>
  )
}
