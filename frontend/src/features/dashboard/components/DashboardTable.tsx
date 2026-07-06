import { useTranslation } from 'react-i18next'
import { Loader2, Pencil, Plus, Trash2 } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import type { Dashboard } from '@/features/dashboard/types'

export function DashboardTable({
  dashboards,
  loading,
  search,
  onSearchChange,
  onSelect,
  onCreate,
  onEdit,
  onDelete,
}: {
  dashboards: Dashboard[]
  loading: boolean
  search: string
  onSearchChange: (value: string) => void
  onSelect: (id: number) => void
  onCreate: () => void
  onEdit: (d: Dashboard) => void
  onDelete: (d: Dashboard) => void
}) {
  const { t } = useTranslation()

  return (
    <div className="flex h-full flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-base font-semibold">{t('dashboards.list.title')}</h1>
        <div className="flex items-center gap-2">
          <Input
            value={search}
            onChange={(e) => onSearchChange(e.target.value)}
            placeholder={t('dashboards.list.searchPlaceholder') ?? ''}
            className="h-9 w-64"
          />
          <Button size="sm" onClick={onCreate}>
            <Plus size={14} className="mr-1" />
            {t('dashboards.list.create')}
          </Button>
        </div>
      </div>

      <div className="min-h-0 flex-1 overflow-x-auto rounded-lg border border-border bg-card">
        {loading && (
          <div className="flex items-center justify-center gap-2 py-10 text-xs text-muted-foreground">
            <Loader2 size={14} className="animate-spin" />
            {t('dashboards.list.loading')}
          </div>
        )}
        {!loading && dashboards.length === 0 && (
          <div className="py-10 text-center text-xs text-muted-foreground">
            {t('dashboards.list.empty')}
          </div>
        )}
        {!loading && dashboards.length > 0 && (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border text-left text-xs uppercase tracking-wide text-muted-foreground">
                <th className="px-3 py-2 font-medium">{t('dashboards.table.col.name')}</th>
                <th className="px-3 py-2 font-medium">{t('dashboards.table.col.description')}</th>
                <th className="px-3 py-2 font-medium">{t('dashboards.table.col.modified')}</th>
                <th className="px-3 py-2 font-medium" />
              </tr>
            </thead>
            <tbody>
              {dashboards.map((d) => (
                <DashboardRow
                  key={d.id}
                  dashboard={d}
                  onSelect={onSelect}
                  onEdit={onEdit}
                  onDelete={onDelete}
                />
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}

function DashboardRow({
  dashboard: d,
  onSelect,
  onEdit,
  onDelete,
}: {
  dashboard: Dashboard
  onSelect: (id: number) => void
  onEdit: (d: Dashboard) => void
  onDelete: (d: Dashboard) => void
}) {
  const { t } = useTranslation()
  const modified = d.modifiedDate ? new Date(d.modifiedDate).toLocaleString() : '—'
  return (
    <tr
      onClick={() => onSelect(d.id)}
      className="cursor-pointer border-b border-border/60 last:border-0 hover:bg-muted/40"
    >
      <td className="px-3 py-2 font-medium text-foreground">{d.name}</td>
      <td className="px-3 py-2 text-muted-foreground">{d.description || '—'}</td>
      <td className="px-3 py-2 text-xs text-muted-foreground">{modified}</td>
      <td className="px-3 py-2 text-right" onClick={(e) => e.stopPropagation()}>
        {!d.systemOwner && (
          <div className="inline-flex items-center gap-1">
            <button
              type="button"
              onClick={() => onEdit(d)}
              className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
              aria-label={t('dashboards.actions.edit') ?? 'Edit'}
              title={t('dashboards.actions.edit') ?? 'Edit'}
            >
              <Pencil size={14} />
            </button>
            <button
              type="button"
              onClick={() => onDelete(d)}
              className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
              aria-label={t('dashboards.list.delete') ?? 'Delete'}
              title={t('dashboards.list.delete') ?? 'Delete'}
            >
              <Trash2 size={14} />
            </button>
          </div>
        )}
      </td>
    </tr>
  )
}
