import { useTranslation } from 'react-i18next'
import { useDatasets } from '@/features/dashboard/hooks/useDatasets'

export function DatasetSelect({
  value,
  onChange,
}: {
  value: string
  onChange: (dataset: string) => void
}) {
  const { t } = useTranslation()
  const query = useDatasets()
  const datasets = query.data ?? []

  return (
    <select
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className="h-9 w-full rounded-md border border-input bg-background px-2 text-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
    >
      <option value="">
        {query.isLoading
          ? t('dashboards.editor.dataset.loading')
          : t('dashboards.editor.dataset.placeholder')}
      </option>
      {datasets.map((d) => (
        <option key={d} value={d}>
          {d}
        </option>
      ))}
    </select>
  )
}
