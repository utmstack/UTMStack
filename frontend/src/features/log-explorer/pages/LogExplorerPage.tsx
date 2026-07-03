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

  // Deep-link seed: any query param becomes a filter (?field=value; comma-list
  // ?field=a,b,c → IS_ONE_OF_TERMS). Reserved: ?@timestamp=<preset-id> sets the
  // range (e.g. 24h, 7d, 30d). No params → normal open, existing tabs shown.
  useEffect(() => {
    if (seededRef.current) return
    const params = new URLSearchParams(location.search)

    let range = presetRange('30d')
    const filters: FilterType[] = []
    for (const [field, raw] of params.entries()) {
      if (field === TS_FIELD) {
        range = presetRange(raw)
        continue
      }
      const values = raw.split(',').map((s) => s.trim()).filter(Boolean)
      if (values.length === 0) continue
      filters.push(
        values.length === 1
          ? { field, operator: 'IS', value: values[0] }
          : { field, operator: 'IS_ONE_OF_TERMS', value: values }
      )
    }

    if (filters.length === 0) return

    seededRef.current = true
    const first = filters[0]
    const name =
      filters.length === 1
        ? `${first.field}: ${Array.isArray(first.value) ? first.value.join(',') : String(first.value)}`
        : 'Filtered'
    tabs.addTab({
      name,
      patternStr: 'v11-log-*',
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
