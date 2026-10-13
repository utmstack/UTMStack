import { useState } from 'react'
import { useTranslation, Trans } from 'react-i18next'
import { Cloud, Loader2, ShieldCheck } from 'lucide-react'
import { Section } from '@/features/integrations/components/ui/Section'
import { FlowNode, FlowEdge } from '../ForwarderGuide'
import { CloudTenantList } from '@/features/integrations/components/setup/cloud/CloudTenantList'
import { CloudTenantForm } from '@/features/integrations/components/setup/cloud/CloudTenantForm'
import { useIntegrations } from '@/features/integrations/hooks/useIntegrations'
import type { CloudConfigField } from '@/features/integrations/components/setup/cloud/builders/cloudGuideBuilder'
import type { Integration, ConfigGroupResponse } from '@/features/integrations/types'

const O365_ROOT = 'integrations.setup.collector.office365'
const GRAPH_ROOT = 'integrations.setup.collector.o365_graph'
const O365_IMG = '/integrations/guides/collector/office365'
const GRAPH_IMG = '/integrations/guides/collector/o365_graph'

const O365_FIELDS: CloudConfigField[] = [
  { key: 'office365_tenant_id',         labelKey: `${O365_ROOT}.fields.tenantId.label`,     placeholder: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx' },
  { key: 'office365_client_id',         labelKey: `${O365_ROOT}.fields.clientId.label`,     placeholder: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx' },
  { key: 'office365_client_secret',     labelKey: `${O365_ROOT}.fields.clientSecret.label`, secret: true, type: 'password' },
  { key: 'office365_cloud_environment', labelKey: `${O365_ROOT}.fields.cloudEnv.label`,     placeholder: 'Commercial' },
]

export const O365_GRAPH_FIELDS: CloudConfigField[] = [
  { key: 'o365_graph_tenant_id',         labelKey: `${GRAPH_ROOT}.fields.tenantId.label`,     placeholder: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx' },
  { key: 'o365_graph_client_id',         labelKey: `${GRAPH_ROOT}.fields.clientId.label`,     placeholder: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx' },
  { key: 'o365_graph_client_secret',     labelKey: `${GRAPH_ROOT}.fields.clientSecret.label`, secret: true, type: 'password' },
  { key: 'o365_graph_cloud_environment', labelKey: `${GRAPH_ROOT}.fields.cloudEnv.label`,     placeholder: 'Commercial' },
]

const O365_STEPS = [
  { key: 'azurePortal',          img: `${O365_IMG}/o365-portal.png` },
  { key: 'appRegistrations',     img: `${O365_IMG}/o365-app-registration.png` },
  { key: 'registerApp',          img: `${O365_IMG}/o365-register-app.png` },
  { key: 'clientSecret',         img: `${O365_IMG}/o365-certificate-secret.png` },
  { key: 'apiPermissions',       img: `${O365_IMG}/o365-api-permission.png` },
  { key: 'activityFeed',         img: `${O365_IMG}/o365-api-permission-request.png` },
  { key: 'graphPermissions',     img: `${O365_IMG}/o365-activity-feed.png` },
  { key: 'delegatedPermissions', img: `${O365_IMG}/o365-delegated-permission.png` },
  { key: 'appPermissions',       img: `${O365_IMG}/o365-app-permission.png` },
  { key: 'noteIds',              img: `${O365_IMG}/o365-overview-client.png` },
]

// Graph guide reuses the same app-registration flow but points at the
// o365_graph screenshots and Graph-specific permission steps.
const GRAPH_STEPS = [
  { key: 'azurePortal',       img: `${GRAPH_IMG}/graph-portal.png` },
  { key: 'appRegistrations',  img: `${GRAPH_IMG}/graph-app-registrations.png` },
  { key: 'registerApp',       img: `${GRAPH_IMG}/graph-register-app.png` },
  { key: 'clientSecret',      img: `${GRAPH_IMG}/graph-certificate-secret.png` },
  { key: 'apiPermissions',    img: `${GRAPH_IMG}/graph-api-permission.png` },
  { key: 'graphPermissions',  img: `${GRAPH_IMG}/graph-permissions-list.png` },
  { key: 'adminConsent',      img: `${GRAPH_IMG}/graph-admin-consent.png` },
  { key: 'noteIds',           img: `${GRAPH_IMG}/graph-overview-client.png` },
]

function StepImg({ src }: { src: string }) {
  return (
    <img
      src={src}
      alt=""
      className="mt-3 w-full rounded-md border border-border"
      onError={(e) => { e.currentTarget.style.display = 'none' }}
    />
  )
}

function SectionLabel({ title }: { title: string }) {
  return (
    <div className="flex items-center gap-3">
      <div className="h-px flex-1 bg-border" />
      <span className="text-[11px] font-semibold uppercase tracking-widest text-muted-foreground">{title}</span>
      <div className="h-px flex-1 bg-border" />
    </div>
  )
}

export function M365Guide({ module: integration, fields = O365_FIELDS, variant = 'o365' }: { module: Integration; fields?: CloudConfigField[]; variant?: 'o365' | 'graph' }) {
  const { t } = useTranslation()
  const { configGroups: tenantsQuery, deleteConfigGroup } = useIntegrations()

  const ROOT = variant === 'graph' ? GRAPH_ROOT : O365_ROOT
  const STEPS = variant === 'graph' ? GRAPH_STEPS : O365_STEPS
  const moduleName = integration.moduleName ?? (variant === 'graph' ? 'O365_GRAPH' : 'O365')
  const tenantList = tenantsQuery(moduleName)
  const tenants: ConfigGroupResponse[] = tenantList.data ?? []

  const [editing, setEditing] = useState<ConfigGroupResponse | null>(null)

  const handleDelete = (name: string) => {
    deleteConfigGroup.mutate({ integration: moduleName, name })
    if (editing?.name === name) setEditing(null)
  }

  return (
    <div className="space-y-4">

      {/* ── Intro + flow diagram ─────────────────────────────────── */}
      <Section title={t(`${ROOT}.intro.title`)}>
        <p className="text-sm text-foreground/90">{t(`${ROOT}.intro.body`)}</p>
        <div className="mt-3 flex items-stretch justify-center gap-1 sm:gap-2">
          <FlowNode
            icon={<Cloud size={20} className="text-muted-foreground" />}
            title="Microsoft 365"
            sub={t(`${ROOT}.intro.diagram.source`)}
            tone="neutral"
          />
          <FlowEdge label={t(`${ROOT}.intro.diagram.edge`)} />
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

      {/* ── Azure app registration steps ─────────────────────────── */}
      <SectionLabel title={t(`${ROOT}.appSection`)} />

      {STEPS.map((s, idx) => (
        <Section key={s.key} title={t(`${ROOT}.steps.${s.key}.title`)} step={idx + 1}>
          <p className="text-sm text-foreground/90">
            <Trans
              i18nKey={`${ROOT}.steps.${s.key}.body`}
              components={{ hl: <strong className="font-semibold text-primary" /> }}
            />
          </p>
          {s.key === 'clientSecret' && (
            <p className="mt-2 rounded-md bg-amber-500/10 border border-amber-500/30 px-3 py-2 text-[11px] text-amber-700 dark:text-amber-400">
              {t(`${ROOT}.steps.clientSecret.copyNote`)}
            </p>
          )}
          <StepImg src={s.img} />
        </Section>
      ))}

      {/* ── Credentials form ─────────────────────────────────────── */}
      <SectionLabel title={t(`${ROOT}.credentialsSection`)} />

      <Section title={t(`${ROOT}.credentials.title`)}>
        <p className="mb-3 text-sm text-foreground/90">{t(`${ROOT}.credentials.body`)}</p>
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
            fields={fields}
            editing={editing}
            onCancel={() => setEditing(null)}
            onSaved={() => setEditing(null)}
          />
        </div>
      </Section>

    </div>
  )
}
