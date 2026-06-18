import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Bird, Loader2, Rss, ShieldCheck } from 'lucide-react'
import { Section } from '@/features/integrations/components/ui/Section'
import { FlowNode, FlowEdge } from '../collector/ForwarderGuide'
import { CloudTenantList } from './CloudTenantList'
import { CloudTenantForm } from './CloudTenantForm'
import { useIntegrations } from '@/features/integrations/hooks/useIntegrations'
import { CROWDSTRIKE_FIELDS } from './builders/crowdstrike'
import type { Integration, TenantResponse } from '@/features/integrations/types'

const ROOT = 'integrations.setup.cloud.crowdstrike'
const IMG = '/integrations/guides/crowdstrike'

// Numbered steps mirror the legacy CrowdStrike guide 1:1 (API Clients & Keys →
// create client → scopes/create → record credentials). Images come from the
// legacy guide assets.
const STEPS: Array<{ key: string; img: string }> = [
  { key: 'apiClients', img: `${IMG}/1.png` },
  { key: 'createClient', img: `${IMG}/2.png` },
  { key: 'scopes', img: `${IMG}/3.png` },
  { key: 'recordCredentials', img: `${IMG}/4.png` },
]

// Field → step reference shown above the form (order matches CROWDSTRIKE_FIELDS).
const MAPPING_KEYS = ['appName', 'clientId', 'clientSecret', 'cloudRegionUrl'] as const

export function CrowdStrikeGuide({ integration }: { integration: Integration }) {
  const { t } = useTranslation()
  const { tenants: tenantsQuery, deleteTenant } = useIntegrations()

  const moduleName = integration.moduleName ?? ''
  const tenantList = tenantsQuery(moduleName)
  const tenants: TenantResponse[] = tenantList.data ?? []

  const [editing, setEditing] = useState<TenantResponse | null>(null)

  const handleDelete = (name: string) => {
    deleteTenant.mutate({ moduleName, name })
    if (editing?.name === name) setEditing(null)
  }

  return (
    <div className="space-y-4">
      {/* Intro + flow diagram. CrowdStrike is a puller running inside UTMStack:
          it reads the Falcon Event Streams API directly — nothing is installed. */}
      <Section title={t(`${ROOT}.intro.title`)}>
        <p className="text-sm text-foreground/90">{t(`${ROOT}.intro.body`)}</p>
        <div className="mt-3 flex items-stretch justify-center gap-1 sm:gap-2">
          <FlowNode
            icon={<Bird size={20} className="text-muted-foreground" />}
            title="Falcon"
            sub={t(`${ROOT}.intro.diagram.source`)}
            tone="neutral"
          />
          <FlowEdge label={t(`${ROOT}.intro.diagram.edge1`)} />
          <FlowNode
            icon={<Rss size={20} />}
            title={t(`${ROOT}.intro.diagram.api`)}
            sub={t(`${ROOT}.intro.diagram.apiSub`)}
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

      {/* Numbered CrowdStrike API client setup steps. */}
      {STEPS.map((s, idx) => (
        <Section key={s.key} title={t(`${ROOT}.sections.${s.key}.title`)} step={idx + 1}>
          <p className="text-sm text-foreground/90">{t(`${ROOT}.sections.${s.key}.body`)}</p>
          <img
            src={s.img}
            alt=""
            className="mt-3 max-w-full rounded-md border border-border"
            onError={(e) => {
              e.currentTarget.style.display = 'none'
            }}
          />
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
              isDeleting={deleteTenant.isPending}
              onEdit={setEditing}
              onDelete={handleDelete}
            />
          )}
          <CloudTenantForm
            moduleName={moduleName}
            fields={CROWDSTRIKE_FIELDS}
            editing={editing}
            onCancel={() => setEditing(null)}
            onSaved={() => setEditing(null)}
          />
        </div>
      </Section>
    </div>
  )
}
