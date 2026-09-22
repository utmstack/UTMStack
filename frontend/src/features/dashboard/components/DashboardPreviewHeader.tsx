import type { ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { ArrowLeft, Pencil, Sparkles, Trash2 } from 'lucide-react'
import type { Dashboard } from '@/features/dashboard/types'

export function DashboardPreviewHeader({
  dashboard,
  onBack,
  onEdit,
  onEditWithAi,
  onDelete,
  right,
}: {
  dashboard: Dashboard
  onBack: () => void
  /** Enter widget-layout edit mode. Omit while already editing or for systemOwner dashboards. */
  onEdit?: () => void
  /** Open the assistant, scoped to this dashboard. Omit while layout-editing, for systemOwner dashboards, or when SOC-AI isn't configured. */
  onEditWithAi?: () => void
  /** Delete (tenant-owned) or dismiss (system-owned) — the backend decides which; always available. */
  onDelete: (d: Dashboard) => void
  /** Consumption controls (time range) or the editor bar while editing. */
  right?: ReactNode
}) {
  const { t } = useTranslation()
  const canModify = !dashboard.systemOwner

  return (
    <div className="flex items-center gap-3">
      <button
        type="button"
        onClick={onBack}
        className="inline-flex items-center gap-1.5 rounded-md px-2 py-1 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
      >
        <ArrowLeft size={15} />
        {t('dashboards.tabs.dashboards')}
      </button>
      <span className="text-border">/</span>
      <div className="min-w-0 flex-1">
        <h1 className="truncate text-base font-semibold">{dashboard.name}</h1>
        {dashboard.description && (
          <p className="truncate text-xs text-muted-foreground">{dashboard.description}</p>
        )}
      </div>

      <div className="flex shrink-0 items-center gap-2">
        {right}
        {canModify && onEditWithAi && (
          <button
            type="button"
            onClick={onEditWithAi}
            className="inline-flex h-9 items-center gap-1.5 rounded-md border border-primary/30 bg-primary/5 px-3 text-sm font-medium text-primary transition-colors hover:bg-primary/10"
          >
            <Sparkles size={14} />
            {t('dashboards.actions.editWithAi')}
          </button>
        )}
        {canModify && onEdit && (
          <button
            type="button"
            onClick={onEdit}
            className="inline-flex h-9 w-9 items-center justify-center rounded-md border border-border text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
            aria-label={t('dashboards.actions.edit') ?? 'Edit'}
            title={t('dashboards.actions.edit') ?? 'Edit'}
          >
            <Pencil size={15} />
          </button>
        )}
        <button
          type="button"
          onClick={() => onDelete(dashboard)}
          className="inline-flex h-9 w-9 items-center justify-center rounded-md border border-border text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
          aria-label={t('dashboards.list.delete') ?? 'Delete'}
          title={t('dashboards.list.delete') ?? 'Delete'}
        >
          <Trash2 size={15} />
        </button>
      </div>
    </div>
  )
}
