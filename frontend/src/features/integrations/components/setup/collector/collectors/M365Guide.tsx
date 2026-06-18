import { useState } from 'react'
import { useTranslation, Trans } from 'react-i18next'
import { Cloud, Loader2, ShieldCheck } from 'lucide-react'
import { Section } from '@/features/integrations/components/ui/Section'
import { FlowNode, FlowEdge } from '../ForwarderGuide'
import { CloudTenantList } from '@/features/integrations/components/setup/cloud/CloudTenantList'
import { CloudTenantForm } from '@/features/integrations/components/setup/cloud/CloudTenantForm'
import { useIntegrations } from '@/features/integrations/hooks/useIntegrations'
import type { CloudConfigField } from '@/features/integrations/components/setup/cloud/builders/cloudGuideBuilder'
import type { Integration, TenantResponse } from '@/features/integrations/types'

const ROOT = 'integrations.setup.collector.office365'
const IMG = '/integrations/guides/collector/office365'

const FIELDS: CloudConfigField[] = [
  { key: 'office365_tenant_id',         labelKey: `${ROOT}.fields.tenantId.label`,     placeholder: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx' },
  { key: 'office365_client_id',         labelKey: `${ROOT}.fields.clientId.label`,     placeholder: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx' },
  { key: 'office365_client_secret',     labelKey: `${ROOT}.fields.clientSecret.label`, secret: true, type: 'password' },
  { key: 'office365_cloud_environment', labelKey: `${ROOT}.fields.cloudEnv.label`,     placeholder: 'Commercial' },
]

const STEPS = [
  { key: 'azurePortal',          img: `${IMG}/o365-portal.png` },
  { key: 'appRegistrations',     img: `${IMG}/o365-app-registration.png` },
  { key: 'registerApp',          img: `${IMG}/o365-register-app.png` },
  { key: 'clientSecret',         img: `${IMG}/o365-certificate-secret.png` },
  { key: 'apiPermissions',       img: `${IMG}/o365-api-permission.png` },
  { key: 'activityFeed',         img: `${IMG}/o365-api-permission-request.png` },
  { key: 'graphPermissions',     img: `${IMG}/o365-activity-feed.png` },
  { key: 'delegatedPermissions', img: `${IMG}/o365-delegated-permission.png` },
  { key: 'appPermissions',       img: `${IMG}/o365-app-permission.png` },
  { key: 'noteIds',              img: `${IMG}/o365-overview-client.png` },
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

export function M365Guide({ module: integration }: { module: Integration }) {
  const { t } = useTranslation()
  const { tenants: tenantsQuery, deleteTenant } = useIntegrations()

  const moduleName = integration.moduleName ?? 'O365'
  const tenantList = tenantsQuery(moduleName)
  const tenants: TenantResponse[] = tenantList.data ?? []

  const [editing, setEditing] = useState<TenantResponse | null>(null)

  const handleDelete = (name: string) => {
    deleteTenant.mutate({ moduleName, name })
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
              isDeleting={deleteTenant.isPending}
              onEdit={setEditing}
              onDelete={handleDelete}
            />
          )}
          <CloudTenantForm
            moduleName={moduleName}
            fields={FIELDS}
            editing={editing}
            onCancel={() => setEditing(null)}
            onSaved={() => setEditing(null)}
          />
        </div>
      </Section>

    </div>
  )
}
