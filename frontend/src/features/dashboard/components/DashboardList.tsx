import { useTranslation } from 'react-i18next'
import { Loader2, Pencil, Plus, Trash2 } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import type { Dashboard } from '@/features/dashboard/types'

export function DashboardList({
  dashboards,
  selectedId,
  search,
  loading,
  onSearchChange,
  onSelect,
  onCreate,
  onRename,
  onDelete,
}: {
  dashboards: Dashboard[]
  selectedId: number | null
  search: string
  loading: boolean
  onSearchChange: (value: string) => void
  onSelect: (id: number) => void
  onCreate: () => void
  onRename: (d: Dashboard) => void
  onDelete: (d: Dashboard) => void
}) {
  const { t } = useTranslation()

  return (
    <aside className="flex h-full w-72 shrink-0 flex-col overflow-hidden rounded-lg border border-border bg-card">
      <div className="flex items-center justify-between gap-2 border-b border-border px-3 py-2">
        <span className="text-sm font-semibold">{t('dashboards.list.title')}</span>
        <Button variant="outline" size="sm" onClick={onCreate}>
          <Plus size={14} className="mr-1" />
          {t('dashboards.list.create')}
        </Button>
      </div>
      <div className="border-b border-border px-3 py-2">
        <Input
          value={search}
          onChange={(e) => onSearchChange(e.target.value)}
          placeholder={t('dashboards.list.searchPlaceholder') ?? ''}
          className="h-9"
        />
      </div>
      <div className="min-h-0 flex-1 overflow-y-auto">
        {loading && (
          <div className="flex items-center justify-center gap-2 px-3 py-6 text-xs text-muted-foreground">
            <Loader2 size={14} className="animate-spin" />
            {t('dashboards.list.loading')}
          </div>
        )}
        {!loading && dashboards.length === 0 && (
          <div className="px-3 py-6 text-center text-xs text-muted-foreground">
            {t('dashboards.list.empty')}
          </div>
        )}
        <ul className="space-y-1 px-2 py-2">
          {dashboards.map((d) => {
            const active = d.id === selectedId
            return (
              <li key={d.id}>
                <div
                  className={cn(
                    'group flex items-center gap-2 rounded-md px-2 py-2 text-sm transition-colors',
                    active
                      ? 'bg-primary/15 text-primary'
                      : 'text-foreground/80 hover:bg-muted'
                  )}
                >
                  <button
                    type="button"
                    onClick={() => onSelect(d.id)}
                    className="flex min-w-0 flex-1 flex-col text-left"
                  >
                    <span className="truncate font-medium">{d.name}</span>
                    {d.description && (
                      <span className="truncate text-xs text-muted-foreground">{d.description}</span>
                    )}
                  </button>
                  {!d.systemOwner && (
                    <div className="flex shrink-0 items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100">
                      <button
                        type="button"
                        onClick={() => onRename(d)}
                        className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
                        aria-label={t('dashboards.list.rename') ?? 'Rename'}
                      >
                        <Pencil size={13} />
                      </button>
                      <button
                        type="button"
                        onClick={() => onDelete(d)}
                        className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
                        aria-label={t('dashboards.list.delete') ?? 'Delete'}
                      >
                        <Trash2 size={13} />
                      </button>
                    </div>
                  )}
                </div>
              </li>
            )
          })}
        </ul>
      </div>
    </aside>
  )
}
