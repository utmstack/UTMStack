import { useEffect, useRef } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { presetRange } from '@/shared/components/ui/time-range-picker'
import { LogExplorerTabsBar } from '../components/LogExplorerTabsBar'
import { LogExplorerView } from '../components/LogExplorerView'
import { useLogExplorerTabs } from '../hooks/useLogExplorerTabs'

export function LogExplorerPage() {
  const location = useLocation()
  const navigate = useNavigate()
  const tabs = useLogExplorerTabs()
  const seededRef = useRef(false)

  // Integration "View events" deep-link (?dataType=<value>) → open in a NEW tab.
  useEffect(() => {
    if (seededRef.current) return
    const dataType = new URLSearchParams(location.search).get('dataType')
    if (!dataType) return
    seededRef.current = true
    tabs.addTab({
      name: `dataType: ${dataType}`,
      patternStr: 'v11-log-*',
      filters: [{ field: 'dataType', operator: 'IS', value: dataType }],
      range: presetRange('30d'),
      columns: ['dataType'],
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
