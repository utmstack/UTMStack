import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Cloud, ExternalLink, Loader2, Rss, ShieldCheck, TriangleAlert } from 'lucide-react'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import { FlowNode, FlowEdge } from '../collector/ForwarderGuide'
import { CloudTenantList } from './CloudTenantList'
import { CloudTenantForm } from './CloudTenantForm'
import { useIntegrations } from '@/features/integrations/hooks/useIntegrations'
import { AZURE_FIELDS } from './builders/azure'
import type { Integration, ConfigGroupResponse } from '@/features/integrations/types'

const ROOT = 'integrations.setup.cloud.azure'
const IMG = '/integrations/guides/azure'

// Every numbered step from the legacy Azure guide, in order. A step may carry an
// image, an external Azure doc link, a sample connection string, and/or a warning
// callout — rendered only when present.
interface AzureStep {
  key: string
  img?: string
  doc?: string
  sample?: string
  warn?: boolean
}

const STEPS: AzureStep[] = [
  { key: 'createEventHub', doc: 'https://learn.microsoft.com/azure/event-hubs/event-hubs-create' },
  { key: 'sharedAccessPolicy', img: `${IMG}/event-hub-access-policy.png` },
  {
    key: 'connectionString',
    img: `${IMG}/access-shared-policy.png`,
    sample:
      'Endpoint=sb://utmstacksharedaccesspolicy.servicebus.windows.net/;SharedAccessKeyName=UTMStackSharedAccesspolicy;SharedAccessKey=A1xFRWsEKcS19gGPEykezcVsK4qLAcQ2K+AEhCyITzU=',
  },
  { key: 'consumerGroup', img: `${IMG}/consumer-group.png`, warn: true },
  { key: 'createStorageAccount', doc: 'https://learn.microsoft.com/azure/storage/common/storage-account-create' },
  { key: 'storageContainer', img: `${IMG}/container-name.png` },
  {
    key: 'storageConnection',
    img: `${IMG}/account-storage.png`,
    sample:
      'DefaultEndpointsProtocol=https;AccountName=utmstackstorageaccount;AccountKey=ETOPnkd/hDAWidkEpPZDiXffQPku/SZdXhPSLnfqdRTalssdEuPkZwIcouzXjCLb/xPZjzhmHfwRCGo0SBSw==;EndpointSuffix=core.windows.net',
  },
  { key: 'eventGrid', img: `${IMG}/event-grid.png` },
  { key: 'createEventSubscription', img: `${IMG}/event-subs.png` },
  { key: 'configureEventSubscription', img: `${IMG}/event-hub-selection.png` },
]

const MAPPING_KEYS = ['eventHubConnection', 'consumerGroup', 'storageContainer', 'storageConnection'] as const

export function AzureGuide({ integration }: { integration: Integration }) {
  const { t } = useTranslation()
  const { configGroups: tenantsQuery, deleteConfigGroup } = useIntegrations()

  const moduleName = integration.moduleName ?? ''
  const tenantList = tenantsQuery(moduleName)
  const tenants: ConfigGroupResponse[] = tenantList.data ?? []

  const [editing, setEditing] = useState<ConfigGroupResponse | null>(null)

  const handleDelete = (name: string) => {
    deleteConfigGroup.mutate({ integration: moduleName, name })
    if (editing?.name === name) setEditing(null)
  }

  return (
    <div className="space-y-4">
      {/* Intro + flow diagram. Azure is a puller running inside UTMStack: the
          subscription streams events into an Event Hub and UTMStack reads it. */}
      <Section title={t(`${ROOT}.intro.title`)}>
        <p className="text-sm text-foreground/90">{t(`${ROOT}.intro.body`)}</p>
        <div className="mt-3 flex items-stretch justify-center gap-1 sm:gap-2">
          <FlowNode
            icon={<Cloud size={20} className="text-muted-foreground" />}
            title="Azure"
            sub={t(`${ROOT}.intro.diagram.source`)}
            tone="neutral"
          />
          <FlowEdge label={t(`${ROOT}.intro.diagram.edge1`)} />
          <FlowNode
            icon={<Rss size={20} />}
            title={t(`${ROOT}.intro.diagram.eventHub`)}
            sub={t(`${ROOT}.intro.diagram.eventHubSub`)}
            tone="accent"
          />
          <FlowEdge label={t(`${ROOT}.intro.diagram.edge2`)} />
          <FlowNode
            icon={<ShieldCheck size={20} />}
            title="UTMStack"
            sub={t(`${ROOT}.intro.diagram.utmstackSub`)}
            tone="brand"
          />
        </div>
        <p className="mt-3 rounded-md bg-muted/40 px-3 py-2 text-[11px] text-muted-foreground">
          {t(`${ROOT}.intro.note`)}
        </p>
      </Section>

      {/* Numbered Azure setup steps (mirrors the legacy guide 1:1). */}
      {STEPS.map((s, idx) => (
        <Section key={s.key} title={t(`${ROOT}.sections.${s.key}.title`)} step={idx + 1}>
          <p className="text-sm text-foreground/90">{t(`${ROOT}.sections.${s.key}.body`)}</p>

          {s.doc && (
            <a
              href={s.doc}
              target="_blank"
              rel="noreferrer noopener"
              className="mt-2 inline-flex items-center gap-1.5 text-sm font-medium text-primary hover:underline"
            >
              <ExternalLink size={13} />
              {t(`${ROOT}.sections.${s.key}.docLabel`)}
            </a>
          )}

          {s.warn && (
            <div className="mt-3 flex gap-2 rounded-md border border-amber-500/30 bg-amber-500/[0.07] px-3 py-2 text-[11px] text-amber-700 dark:text-amber-300">
              <TriangleAlert size={14} className="mt-0.5 shrink-0" />
              <span>{t(`${ROOT}.sections.${s.key}.warn`)}</span>
            </div>
          )}

          {s.img && (
            <img
              src={s.img}
              alt=""
              className="mt-3 max-w-full rounded-md border border-border"
              onError={(e) => {
                e.currentTarget.style.display = 'none'
              }}
            />
          )}

          {s.sample && (
            <div className="mt-3">
              <p className="mb-1.5 text-[11px] font-medium text-muted-foreground">{t(`${ROOT}.sampleLabel`)}</p>
              <CodeBlock lang="config" code={s.sample} />
            </div>
          )}
        </Section>
      ))}

      {/* Configuration groups: tenant list + add/edit form, with the field map. */}
      <Section title={t(`${ROOT}.credentials.title`)}>
        <p className="mb-2 text-sm text-foreground/90">{t(`${ROOT}.credentials.body`)}</p>

        <div className="mb-3 rounded-md bg-muted/40 px-3 py-2.5">
          <p className="mb-1.5 text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
            {t(`${ROOT}.credentials.mapping.intro`)}
          </p>
          <ul className="space-y-1 text-[11px] text-muted-foreground">
            {MAPPING_KEYS.map((k) => (
              <li key={k} className="flex gap-1.5">
                <span className="text-primary">•</span>
                <span>{t(`${ROOT}.credentials.mapping.${k}`)}</span>
              </li>
            ))}
          </ul>
        </div>

        <div className="space-y-3">
          {tenantList.isLoading ? (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Loader2 className="h-4 w-4 animate-spin" />
              {t('integrations.setup.cloud.tenants.loading')}
            </div>
          ) : (
            <CloudTenantList
              tenants={tenants}
              isDeleting={deleteConfigGroup.isPending}
              onEdit={setEditing}
              onDelete={handleDelete}
            />
          )}
          <CloudTenantForm
            moduleName={moduleName}
            fields={AZURE_FIELDS}
            editing={editing}
            onCancel={() => setEditing(null)}
            onSaved={() => setEditing(null)}
          />
        </div>
      </Section>
    </div>
  )
}
