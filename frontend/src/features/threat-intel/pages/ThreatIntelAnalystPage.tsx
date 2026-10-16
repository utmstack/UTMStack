import { useTranslation } from 'react-i18next'
import { ThreatIntelConsole } from '../console/ThreatIntelConsole'

export function ThreatIntelAnalystPage() {
  const { t } = useTranslation()
  return (
    <div className="flex h-full flex-col px-4 py-4">
      <h1 className="shrink-0 text-lg font-semibold">{t('threatIntel.tabs.analyst')}</h1>
      <div className="mt-3 min-h-0 flex-1 rounded-md border border-border/40 overflow-hidden">
        <ThreatIntelConsole />
      </div>
    </div>
  )
}
