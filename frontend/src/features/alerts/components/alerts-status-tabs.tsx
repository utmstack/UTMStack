import { useTranslation } from 'react-i18next'
import { StatusTabs } from '@/shared/components/ui/status-tabs'
import { STATUS_VALUE, STATUS_TABS, type StatusKey, type StatusTab } from '../types/alert.types'

export function AlertsStatusTabs({
  current,
  counts,
  onChange,
}: {
  current: StatusTab
  counts: Record<string, number>
  onChange: (s: StatusTab) => void
}) {
  const { t } = useTranslation()
  return (
    <StatusTabs
      current={current}
      onChange={onChange}
      tabs={STATUS_TABS.map((id) => ({
        id,
        label: id === 'all' ? t('alerts.statusTabs.all') : t(`alerts.status.${id}`),
        count: id === 'all' ? undefined : counts[STATUS_VALUE[id as StatusKey]],
      }))}
    />
  )
}
