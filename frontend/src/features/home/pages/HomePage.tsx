import { useTranslation } from 'react-i18next'
import { AlertTriangle, Flame, ShieldAlert, Workflow } from 'lucide-react'
import { useActivePlaybooks, useAlertKpis, useOpenIncidents } from '../hooks/use-overview'
import { fmtCount } from '../components/helpers'
import { ChatHero } from '../components/ChatHero'
import { KpiTile } from '../components/KpiTile'
import { DatasourcesCard } from '../components/DatasourcesCard'
import { SystemHealthCard } from '../components/SystemHealthCard'
import { MitreTechniquesCard } from '../components/MitreTechniquesCard'
import { TopAssetsCard } from '../components/TopAssetsCard'
import { ComplianceCard } from '../components/ComplianceCard'
import { useSocAi } from '@/features/soc-ai/SocAiProvider'
import { HomeChatTranscript } from '../components/HomeChatTranscript'

export function HomePage() {
  const { t } = useTranslation()
  const alerts = useAlertKpis()
  const incidents = useOpenIncidents()
  const playbooks = useActivePlaybooks()
  const { homeMessages } = useSocAi()

  return (
    <div className="w-full px-6 py-10">
      <ChatHero />

      {homeMessages.length === 0 ? (
        <div className="mt-10 grid grid-cols-12 gap-4">
        <div className="col-span-6 md:col-span-3">
          <KpiTile
            icon={AlertTriangle}
            label={t('home.kpis.alerts24h')}
            value={fmtCount(alerts.alerts24h)}
            loading={alerts.isLoading}
            sparkline={alerts.sparkline}
            accent="text-red-500"
            href="/threat-management/alerts"
          />
        </div>
        <div className="col-span-6 md:col-span-3">
          <KpiTile
            icon={Flame}
            label={t('home.kpis.openIncidents')}
            value={fmtCount(incidents.count)}
            loading={incidents.isLoading}
            sublabel={t('home.kpis.openIncidentsSub')}
            accent="text-amber-500"
            href="/threat-management/incidents"
          />
        </div>
        <div className="col-span-6 md:col-span-3">
          <KpiTile
            icon={ShieldAlert}
            label={t('home.kpis.highSeverity24h')}
            value={fmtCount(alerts.high24h)}
            loading={alerts.isLoading}
            sublabel={t('home.kpis.highSeveritySub', { n: fmtCount(alerts.alerts24h) })}
            accent="text-rose-500"
            href="/threat-management/alerts"
          />
        </div>
        <div className="col-span-6 md:col-span-3">
          <KpiTile
            icon={Workflow}
            label={t('home.kpis.activePlaybooks')}
            value={fmtCount(playbooks.count)}
            loading={playbooks.isLoading}
            sublabel={t('home.kpis.activePlaybooksSub')}
            accent="text-violet-500"
          />
        </div>

        <div className="col-span-12 lg:col-span-8">
          <DatasourcesCard />
        </div>
        <div className="col-span-12 lg:col-span-4">
          <SystemHealthCard />
        </div>

        <div className="col-span-12 lg:col-span-6">
          <MitreTechniquesCard />
        </div>
        <div className="col-span-12 lg:col-span-6">
          <TopAssetsCard />
        </div>

        <div className="col-span-12">
          <ComplianceCard />
        </div>
        </div>
      ) : (
        <HomeChatTranscript />
      )}
    </div>
  )
}
