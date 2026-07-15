import { useTranslation } from 'react-i18next'
import { Loader2, Plus, Save, X } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'

export function DashboardEditorBar({
  dirty,
  saving,
  onSave,
  onDiscard,
  onAddWidget,
}: {
  dirty: boolean
  saving: boolean
  onSave: () => void
  onDiscard: () => void
  // Always shown here rather than inside the grid — a card appended at the
  // bottom of a growing single-column list of widgets would require
  // scrolling past everything to reach it.
  onAddWidget: () => void
}) {
  const { t } = useTranslation()
  return (
    <div className="flex flex-wrap items-center gap-2 rounded-lg border border-border bg-card px-3 py-2">
      <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
        {t('dashboards.editor.editing')}
      </span>
      <div className="ml-auto flex items-center gap-2">
        <Button variant="outline" size="sm" onClick={onAddWidget}>
          <Plus size={14} className="mr-1" />
          {t('dashboards.editor.addWidget')}
        </Button>
        <Button variant="outline" size="sm" onClick={onDiscard} disabled={saving}>
          <X size={14} className="mr-1" />
          {t('dashboards.editor.discard')}
        </Button>
        <Button size="sm" onClick={onSave} disabled={!dirty || saving}>
          {saving ? (
            <Loader2 size={14} className="mr-1 animate-spin" />
          ) : (
            <Save size={14} className="mr-1" />
          )}
          {t('dashboards.editor.save')}
        </Button>
      </div>
    </div>
  )
}
