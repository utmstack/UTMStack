import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Loader2, Search, X } from 'lucide-react'
import { Input } from '@/shared/components/ui/input'
import { Pagination } from '@/shared/components/ui/pagination'
import { useVisualizations } from '@/features/dashboard/hooks/useVisualizations'
import { DEFAULT_PAGE_SIZE } from '@/features/dashboard/constants'
import type { Visualization } from '@/features/dashboard/types'

export function AddVisualizationDrawer({
  open,
  excludedIds,
  onClose,
  onPick,
}: {
  open: boolean
  excludedIds: Set<number>
  onClose: () => void
  onPick: (viz: Visualization) => void
}) {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(0)
  const [size, setSize] = useState(DEFAULT_PAGE_SIZE)

  const query = useVisualizations({ name: search || undefined, page, size })

  if (!open) return null

  const items = query.data?.data ?? []
  const total = query.data?.total ?? 0

  return (
    <div className="fixed inset-0 z-40 flex justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex h-full w-full max-w-md flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-start justify-between gap-4 border-b border-border px-5 py-4">
          <div>
            <h2 className="text-base font-semibold">{t('dashboards.addWidget.title')}</h2>
            <p className="mt-1 text-xs text-muted-foreground">{t('dashboards.addWidget.subtitle')}</p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <X size={16} />
          </button>
        </header>

        <div className="border-b border-border px-5 py-3">
          <div className="relative">
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
              return (
                <li key={viz.id}>
                  <button
                    type="button"
                    disabled={already}
                    onClick={() => onPick(viz)}
                    className="flex w-full flex-col items-start gap-1 rounded-md border border-transparent px-3 py-2 text-left transition-colors hover:border-border hover:bg-muted disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    <span className="text-sm font-medium">{viz.name}</span>
                    {viz.description && (
                      <span className="text-xs text-muted-foreground">{viz.description}</span>
                    )}
                    {already && (
                      <span className="text-[10px] uppercase tracking-wide text-muted-foreground/70">
                        {t('dashboards.addWidget.alreadyAdded')}
                      </span>
                    )}
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
      </div>
    </div>
  )
}
