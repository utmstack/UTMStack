import { TimeRangePicker, type TimeRange } from '@/shared/components/ui/time-range-picker'

export function DashboardTimeRange({
  value,
  onChange,
}: {
  value: TimeRange
  onChange: (next: TimeRange) => void
}) {
  return <TimeRangePicker value={value} onChange={onChange} align="right" />
}
