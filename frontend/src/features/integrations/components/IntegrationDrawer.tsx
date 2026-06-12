import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { Check, ExternalLink, X } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { LogoTile } from '@/features/integrations/components/ui/LogoTile'
import { KIND_META } from '@/features/integrations/constants'
import { AgentSetup } from '@/features/integrations/components/setup/AgentSetup'
import { CollectorSetup } from '@/features/integrations/components/setup/collector/CollectorSetup'
import { CloudSetup } from '@/features/integrations/components/setup/cloud/CloudSetup'
import { CustomSetup } from '@/features/integrations/components/setup/CustomSetup'
import type { Integration } from '@/features/integrations/types'

interface IntegrationDrawerProps {
  integration: Integration
  onClose: () => void
}

export function IntegrationDrawer({ integration: i, onClose }: IntegrationDrawerProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const km = KIND_META[i.kind]
  const isCollectorGroup = km?.group === 'collectors'

  const scrollToCloudTenants = () => {
    document
      .getElementById('cloud-tenants-section')
      ?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }

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
              <LogoTile src={i.logo} alt={i.name} darkInvert={i.darkInvert} size="sm" />
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                  <span>{i.category}</span>
                  <span>·</span>
                  <span className={km?.tone}>{t(`integrations.kind.${i.kind}`)}</span>
                </div>
                <h2 className="mt-0.5 truncate text-xl font-semibold">{i.name}</h2>
                <div className="mt-2 flex items-center gap-2">
                  {i.status === 'configured' ? (
                    <span className="inline-flex items-center gap-1 rounded-md bg-emerald-500/15 px-1.5 py-0.5 text-[11px] font-medium text-emerald-600 ring-1 ring-emerald-500/30 dark:text-emerald-300">
                      <Check size={11} />
                      {t('integrations.status.configured')}
                    </span>
                  ) : (
                    <span className="rounded-md border border-dashed border-border px-1.5 py-0.5 text-[11px] text-muted-foreground">
                      {t('integrations.status.available')}
                    </span>
                  )}
                  {i.rate ? (
                    <span className="text-[11px] text-muted-foreground">
                      <span className="font-mono">{i.rate}</span> {t('integrations.drawer.eventsRate', { rate: i.rate, count: i.events24h ?? 0 })}
                    </span>
                  ) : null}
                </div>
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
            {i.status === 'configured' ? (
              <>
                <Button
                  size="sm"
                  onClick={i.kind === 'cloud' ? scrollToCloudTenants : undefined}
                >
                  <span className="mr-1.5">⚙️</span>
                  {t('integrations.drawer.manage')}
                </Button>
                <Button size="sm" variant="outline" onClick={viewEvents} disabled={!i.dataType}>
                  <ExternalLink size={13} className="mr-1.5" />
                  {t('integrations.drawer.viewEvents')}
                </Button>
              </>
            ) : (
              <Button
                size="sm"
                onClick={i.kind === 'cloud' ? scrollToCloudTenants : undefined}
              >
                <span className="mr-1.5">+</span>
                {i.kind === 'agents & syslog'
                  ? t('integrations.drawer.getInstallCmd')
                  : isCollectorGroup
                    ? t('integrations.drawer.setupCollector')
                    : i.kind === 'cloud'
                      ? t('integrations.drawer.connect')
                      : t('integrations.drawer.configureCustom')}
              </Button>
            )}
            <Button size="sm" variant="ghost" className="ml-auto">
              <ExternalLink size={13} className="mr-1.5" />
              {t('integrations.drawer.documentation')}
            </Button>
          </div>
        </header>

        <div className="flex-1 overflow-y-auto bg-muted/20 p-6">
          {i.kind === 'agents & syslog' && <AgentSetup integration={i} />}
          {i.kind === 'cloud' && <CloudSetup integration={i} />}
          {i.kind === 'custom' && <CustomSetup />}
          {isCollectorGroup && <CollectorSetup integration={i} />}
        </div>
      </div>
    </div>
  )
}
