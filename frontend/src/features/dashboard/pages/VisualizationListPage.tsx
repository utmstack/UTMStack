import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Loader2, Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import {
  useVisualizationMutations,
  useVisualizations,
} from '@/features/dashboard/hooks/useVisualizations'
import { DEFAULT_PAGE_SIZE } from '@/features/dashboard/constants'
import type { Visualization } from '@/features/dashboard/types'

export function VisualizationListPage() {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')

  const query = useVisualizations({
    name: search || undefined,
    page: 0,
    size: DEFAULT_PAGE_SIZE,
  })
  const items = query.data?.data ?? []
  const total = query.data?.total ?? 0

  const { deleteVisualization } = useVisualizationMutations()

  const handleDelete = (v: Visualization) => {
    if (!window.confirm(t('dashboards.visualizationList.confirmDelete', { name: v.name }))) return
    deleteVisualization.mutate(v.id, {
      onSuccess: () => toast.success(t('dashboards.toast.visualizationDeleted')),
      onError: (err) =>
        toast.error(err.message ?? t('dashboards.toast.visualizationDeleteFailed')),
    })
  }

  return (
    <div className="mx-auto flex h-full w-full max-w-[1400px] flex-col gap-4 px-6 py-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-xl font-semibold">{t('dashboards.visualizationList.title')}</h1>
          <p className="text-sm text-muted-foreground">
            {t('dashboards.visualizationList.subtitle', { count: total })}
          </p>
        </div>
        <Button asChild size="sm">
          <Link to="/dashboards/visualizations/new">
            <Plus size={14} className="mr-1" />
            {t('dashboards.visualizationList.create')}
          </Link>
        </Button>
      </div>

      <div className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4">
        <Input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder={t('dashboards.visualizationList.searchPlaceholder') ?? ''}
          className="h-9 max-w-md"
        />

        <div className="min-h-0 flex-1 overflow-x-auto">
          {query.isLoading && (
            <div className="flex items-center justify-center gap-2 py-10 text-xs text-muted-foreground">
              <Loader2 size={14} className="animate-spin" />
              {t('dashboards.visualizationList.loading')}
            </div>
          )}
          {!query.isLoading && items.length === 0 && (
            <div className="py-10 text-center text-xs text-muted-foreground">
              {t('dashboards.visualizationList.empty')}
            </div>
          )}
          {!query.isLoading && items.length > 0 && (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border text-left text-xs uppercase tracking-wide text-muted-foreground">
                  <th className="px-3 py-2 font-medium">
                    {t('dashboards.visualizationList.col.name')}
                  </th>
                  <th className="px-3 py-2 font-medium">
                    {t('dashboards.visualizationList.col.description')}
                  </th>
                  <th className="px-3 py-2 font-medium">
                    {t('dashboards.visualizationList.col.modified')}
                  </th>
                  <th className="px-3 py-2 font-medium" />
                </tr>
              </thead>
              <tbody>
                {items.map((v) => (
                  <tr
                    key={v.id}
                    className="border-b border-border/60 last:border-0 hover:bg-muted/40"
                  >
                    <td className="px-3 py-2 font-medium">{v.name}</td>
                    <td className="px-3 py-2 text-muted-foreground">{v.description || '—'}</td>
                    <td className="px-3 py-2 text-xs text-muted-foreground">
                      {v.modifiedDate ? new Date(v.modifiedDate).toLocaleString() : '—'}
                    </td>
                    <td className="px-3 py-2 text-right">
                      {!v.systemOwner && (
                        <button
                          type="button"
                          onClick={() => handleDelete(v)}
                          disabled={deleteVisualization.isPending}
                          className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-destructive/10 hover:text-destructive disabled:opacity-50"
                          aria-label={t('dashboards.visualizationList.delete') ?? 'Delete'}
                        >
                          <Trash2 size={14} />
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>
    </div>
  )
}

export default VisualizationListPage
