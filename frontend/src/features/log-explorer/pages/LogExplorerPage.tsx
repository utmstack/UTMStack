import { useEffect, useRef } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { presetRange } from '@/shared/components/ui/time-range-picker'
import { LogExplorerTabsBar } from '../components/LogExplorerTabsBar'
import { LogExplorerView } from '../components/LogExplorerView'
import { useLogExplorerTabs } from '../hooks/useLogExplorerTabs'
import type { FilterType } from '../types/log-explorer.types'

const TS_FIELD = '@timestamp'

export function LogExplorerPage() {
  const location = useLocation()
  const navigate = useNavigate()
  const tabs = useLogExplorerTabs()
  const seededRef = useRef(false)

  // Deep-link seed: filter params become terms filters. Reserved: @timestamp
  // selects a preset, while from/to carry an absolute range.
  useEffect(() => {
    if (seededRef.current) return
    const params = new URLSearchParams(location.search)

    let range = presetRange('30d')
    const filters: FilterType[] = []
    const from = params.get('from')
    const to = params.get('to')
    if (from && to) range = { from, to, interval: 'hour' }
    for (const [field, raw] of params.entries()) {
      if (field === 'from' || field === 'to') continue
      if (field === TS_FIELD) {
        range = presetRange(raw)
        continue
      }
      const values = raw.split(',').map((s) => s.trim()).filter(Boolean)
      if (values.length === 0) continue
      filters.push(
        { field, operator: 'IS_ONE_OF_TERMS', value: values }
      )
    }

    if (filters.length === 0 && !(from && to)) return

    seededRef.current = true
    const first = filters[0]
    const name = !first
      ? 'Time range'
      : filters.length === 1
        ? `${first.field}: ${Array.isArray(first.value) ? first.value.join(',') : String(first.value)}`
        : 'Filtered'
    tabs.addTab({
      name,
      patternStr: null,
      filters,
      range,
      columns: filters.map((f) => f.field),
    })
    navigate(location.pathname, { replace: true })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <div className="flex h-full min-h-0 flex-col">
      <LogExplorerTabsBar
        tabs={tabs.tabs}
        activeId={tabs.activeId}
        onSelect={tabs.setActive}
        onAdd={() => tabs.addTab()}
        onClose={tabs.removeTab}
        onRename={tabs.renameTab}
      />
      <div className="min-h-0 flex-1">
        <LogExplorerView
          key={tabs.activeTab.id}
          initial={tabs.activeTab}
          onConfigChange={tabs.updateActive}
        />
      </div>
    </div>
  )
}
