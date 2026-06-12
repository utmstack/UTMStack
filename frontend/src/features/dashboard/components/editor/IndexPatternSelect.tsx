import { useTranslation } from 'react-i18next'
import { useIndexPatterns } from '@/features/dashboard/hooks/useIndexPatterns'
import type { IndexPattern } from '@/features/dashboard/types'

export function IndexPatternSelect({
  value,
  onChange,
}: {
  value: string
  onChange: (pattern: string, id: number | null) => void
}) {
  const { t } = useTranslation()
  const query = useIndexPatterns()
  const patterns: IndexPattern[] = query.data?.data ?? []

  return (
    <select
      value={value}
      onChange={(e) => {
        const next = e.target.value
        const found = patterns.find((p) => p.pattern === next)
        onChange(next, found?.id ?? null)
      }}
      className="h-9 w-full rounded-md border border-input bg-background px-2 text-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
    >
      <option value="">
        {query.isLoading
          ? t('dashboards.editor.indexPattern.loading')
          : t('dashboards.editor.indexPattern.placeholder')}
      </option>
      {patterns.map((p) => (
        <option key={p.id} value={p.pattern}>
          {p.pattern}
        </option>
      ))}
    </select>
  )
}
