import { useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { Loader2 } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { IndexPatternSelect } from '@/features/dashboard/components/editor/IndexPatternSelect'
import { FieldSelect } from '@/features/dashboard/components/editor/FieldSelect'
import { useAggregatableFields } from '@/features/dashboard/hooks/useAggregatableFields'
import type { DashboardFilterChip } from '@/features/dashboard/types'

/**
 * Editor for a single dashboard filter chip. Add-mode when `initial` is null,
 * edit-mode otherwise. The parent (typically {@link DashboardFilterBar}) hosts
 * this in a popover and receives the completed chip through {@link onSubmit}.
 */
export function DashboardFilterChipEditor({
  initial,
  busy,
  onCancel,
  onSubmit,
}: {
  initial: DashboardFilterChip | null
  busy: boolean
  onCancel: () => void
  onSubmit: (chip: DashboardFilterChip) => void
}) {
  const { t } = useTranslation()
  const [chip, setChip] = useState<DashboardFilterChip>(() =>
    initial ?? {
      id: crypto.randomUUID(),
      field: '',
      label: '',
      indexPattern: '',
      multiple: false,
      searchable: true,
    }
  )

  const { fields, isLoading } = useAggregatableFields(chip.indexPattern || null)

  const update = (patch: Partial<DashboardFilterChip>) =>
    setChip((c) => ({ ...c, ...patch }))

  const valid =
    chip.field.trim() !== '' && chip.label.trim() !== '' && chip.indexPattern.trim() !== ''

  return (
    <div className="flex flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl">
      <header className="border-b border-border px-4 py-3">
        <h2 className="text-sm font-semibold">
          {initial ? t('dashboards.filters.editTitle') : t('dashboards.filters.addTitle')}
        </h2>
      </header>

      <div className="max-h-[70vh] overflow-auto px-4 py-4">
        <div className="grid grid-cols-2 gap-3">
          <Field label={t('dashboards.filters.field.indexPattern')}>
            <IndexPatternSelect
              value={chip.indexPattern}
              onChange={(pattern) => update({ indexPattern: pattern, field: '' })}
            />
          </Field>
          <Field label={t('dashboards.filters.field.field')}>
            <FieldSelect
              value={chip.field || null}
              onChange={(next) => update({ field: next })}
              fields={fields}
              loading={isLoading}
              disabled={!chip.indexPattern}
              placeholder={t('dashboards.filters.field.chooseField') ?? ''}
            />
          </Field>
          <Field label={t('dashboards.filters.field.label')}>
            <Input
              value={chip.label}
              onChange={(e) => update({ label: e.target.value })}
              placeholder={t('dashboards.filters.field.labelPlaceholder') ?? ''}
            />
          </Field>
          <Field label={t('dashboards.filters.field.placeholder')}>
            <Input
              value={chip.placeholder ?? ''}
              onChange={(e) => update({ placeholder: e.target.value || undefined })}
              placeholder={t('dashboards.filters.field.placeholderPlaceholder') ?? ''}
            />
          </Field>
          <div className="col-span-2 flex flex-wrap gap-4 text-xs">
            <label className="flex items-center gap-1.5">
              <input
                type="checkbox"
                checked={chip.multiple}
                onChange={(e) => update({ multiple: e.target.checked })}
              />
              {t('dashboards.filters.field.multiple')}
            </label>
            <label className="flex items-center gap-1.5">
              <input
                type="checkbox"
                checked={chip.searchable}
                onChange={(e) => update({ searchable: e.target.checked })}
              />
              {t('dashboards.filters.field.searchable')}
            </label>
          </div>
        </div>
      </div>

      <footer className="flex items-center justify-end gap-2 border-t border-border bg-muted/40 px-4 py-2">
        <Button variant="outline" size="sm" onClick={onCancel} disabled={busy}>
          {t('dashboards.filters.cancel')}
        </Button>
        <Button size="sm" onClick={() => onSubmit(chip)} disabled={!valid || busy}>
          {busy && <Loader2 size={14} className="mr-1 animate-spin" />}
          {t('dashboards.filters.save')}
        </Button>
      </footer>
    </div>
  )
}

function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="space-y-1">
      <label className="block text-[10px] font-medium uppercase tracking-wide text-muted-foreground">
        {label}
      </label>
      {children}
    </div>
  )
}
