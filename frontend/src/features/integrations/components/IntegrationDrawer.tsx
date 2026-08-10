import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { ExternalLink, Filter, ListChecks, X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { useThemeContext } from '@/app/providers/ThemeProvider'
import { Button } from '@/shared/components/ui/button'
import { LogoTile } from '@/features/integrations/components/ui/LogoTile'
import { KIND_META, ingestMeta, categoryLabel } from '@/features/integrations/constants'
import { AgentSetup } from '@/features/integrations/components/setup/AgentSetup'
import { CollectorSetup } from '@/features/integrations/components/setup/collector/CollectorSetup'
import { CloudSetup } from '@/features/integrations/components/setup/cloud/CloudSetup'
import { AWSGuide } from '@/features/integrations/components/setup/cloud/AWSGuide'
import { AzureGuide } from '@/features/integrations/components/setup/cloud/AzureGuide'
import { GcpGuide } from '@/features/integrations/components/setup/cloud/GcpGuide'
import { CrowdStrikeGuide } from '@/features/integrations/components/setup/cloud/CrowdStrikeGuide'
import { M365Guide } from '@/features/integrations/components/setup/collector/collectors/M365Guide'
import { SophosGuide } from '@/features/integrations/components/setup/collector/collectors/SophosGuide'
import { CustomSetup } from '@/features/integrations/components/setup/CustomSetup'
import type { Integration } from '@/features/integrations/types'

const isAWS = (name: string) =>
  (name ?? '').toLowerCase().replace(/\s/g, '').includes('aws')

const isAzure = (name: string) =>
  (name ?? '').toLowerCase().replace(/\s/g, '').includes('azure')

const isM365 = (i: Integration) =>
  (i.moduleName ?? '').toUpperCase() === 'O365' ||
  (i.name ?? '').toLowerCase().replace(/\s/g, '').includes('microsoft365') ||
  (i.name ?? '').toLowerCase().includes('office365')

const isGCP = (i: Integration) => {
  const n = (i.name ?? '').toLowerCase()
  return (
    (i.moduleName ?? '').toUpperCase() === 'GCP' ||
    n.includes('google') ||
    n.includes('gcp') ||
    n.includes('pubsub')
  )
}

// CrowdStrike is an API puller (xdr/plugin) with tenant config, not a collector —
// it gets its own guide even though its technology category lands in the
// collectors group.
const isCrowdStrike = (i: Integration) =>
  (i.moduleName ?? '').toUpperCase() === 'CROWDSTRIKE'

const isSophos = (i: Integration) =>
  (i.moduleName ?? '').toUpperCase() === 'SOPHOS'

// Agents (Linux/Windows/macOS) install a daemon on the host. They're keyed by
// ingest type, not technology category, so they get the install guide regardless
// of which category group they land in.
const isAgent = (i: Integration) => (i.ingestType ?? '').toLowerCase() === 'agent'

// Custom integrations are identified by is_system = false (NOT by a category — they
// carry a real technology category). They always use the Forwarder-based custom
// guide regardless of which category/group they land in.
const isCustom = (i: Integration) => i.systemOwner === false

interface IntegrationDrawerProps {
  integration: Integration
  onClose: () => void
}

export function IntegrationDrawer({ integration: i, onClose }: IntegrationDrawerProps) {
  const { t } = useTranslation()
  const { theme } = useThemeContext()
  const navigate = useNavigate()
  const km = KIND_META[i.kind]
  const logo = theme === 'dark' && i.logoDark ? i.logoDark : i.logo
  const isCollectorGroup = km?.group === 'collectors'
  const active = i.status === 'configured'
  // Deep-link to the rules/pipelines pages pre-filtered by this integration's dataType.
  const dtQuery = i.dataType ? `?dataType=${encodeURIComponent(i.dataType)}` : ''
  const rulesHref = `/alerting-rules${dtQuery}`
  const filtersHref = `/pipelines${dtQuery}`

  const viewEvents = () => {
    if (!i.dataType) return
    navigate(`/log-explorer?dataType=${encodeURIComponent(i.dataType)}`)
  }

  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-[820px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="border-b border-border px-6 py-5">
          <div className="flex items-start justify-between gap-4">
            <div className="flex min-w-0 flex-1 items-start gap-3">
              <LogoTile src={logo} alt={i.name} darkInvert={i.darkInvert} size="sm" />
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                  {i.category && <span>{categoryLabel(i.category)}</span>}
                  {i.category && i.ingestType && <span>·</span>}
                  {i.ingestType && (
                    <span
                      className={cn(
                        'rounded px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wide ring-1',
                        ingestMeta(i.ingestType).pill,
                      )}
                    >
                      {ingestMeta(i.ingestType).label}
                    </span>
                  )}
                </div>
                <h2 className="mt-0.5 truncate text-xl font-semibold">{i.name}</h2>
                {i.rate ? (
                  <div className="mt-2 flex items-center gap-2">
                    <span className="text-[11px] text-muted-foreground">
                      <span className="font-mono">{i.rate}</span> {t('integrations.drawer.eventsRate', { rate: i.rate, count: i.events24h ?? 0 })}
                    </span>
                  </div>
                ) : null}
              </div>
            </div>
            <button
              onClick={onClose}
              className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <X size={16} />
            </button>
          </div>

          <div className="mt-4 flex flex-wrap items-center gap-2">
            {active && (
              <Button size="sm" variant="outline" onClick={viewEvents} disabled={!i.dataType}>
                <ExternalLink size={13} className="mr-1.5" />
                {t('integrations.drawer.viewEvents')}
              </Button>
            )}
            <Button
              size="sm"
              variant="ghost"
              className="ml-auto"
              onClick={() => navigate(rulesHref)}
            >
              <ListChecks size={13} className="mr-1.5" />
              {t('integrations.drawer.manageRules')}
            </Button>
            <Button
              size="sm"
              variant="ghost"
              onClick={() => navigate(filtersHref)}
            >
              <Filter size={13} className="mr-1.5" />
              {t('integrations.drawer.manageFilters')}
            </Button>
          </div>
        </header>

        {/* The setup guide is how an integration gets configured, so it cannot be
            gated on being configured already — that had no way out once the
            enable switch was removed. */}
        <div className="flex-1 overflow-y-auto bg-muted/20 p-6">
          <>
              {isCustom(i) && <CustomSetup integration={i} />}
              {!isCustom(i) && isAgent(i) && <AgentSetup integration={i} />}
              {!isCustom(i) && i.kind === 'cloud' && isAWS(i.name) && <AWSGuide integration={i} />}
              {!isCustom(i) && i.kind === 'cloud' && isAzure(i.name) && <AzureGuide integration={i} />}
              {!isCustom(i) && i.kind === 'cloud' && isGCP(i) && <GcpGuide integration={i} />}
              {!isCustom(i) && i.kind === 'cloud' && isM365(i) && <M365Guide module={i} />}
              {!isCustom(i) && i.kind === 'cloud' && !isAWS(i.name) && !isAzure(i.name) && !isGCP(i) && !isM365(i) && <CloudSetup integration={i} />}
              {!isCustom(i) && isCrowdStrike(i) && <CrowdStrikeGuide integration={i} />}
              {!isCustom(i) && isSophos(i) && <SophosGuide module={i} />}
              {!isCustom(i) && isCollectorGroup && !isCrowdStrike(i) && !isSophos(i) && !isAgent(i) && <CollectorSetup integration={i} />}
          </>
        </div>
      </div>
    </div>
  )
}
