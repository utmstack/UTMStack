import { useTranslation } from 'react-i18next'
import { ThreatIntelConsole } from '../console/ThreatIntelConsole'

export function ThreatIntelAnalystPage() {
  const { t } = useTranslation()
  return (
    <div className="flex h-full flex-col px-4 py-4">
      <h1 className="shrink-0 text-lg font-semibold">{t('threatIntel.tabs.analyst')}</h1>
      <div className="mt-4 min-h-0 flex-1 overflow-hidden rounded-xl border border-border bg-card">
        <ThreatIntelConsole />
      </div>
    </div>
  )
}
