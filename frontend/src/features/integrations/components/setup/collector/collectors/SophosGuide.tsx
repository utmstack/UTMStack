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

const ROOT = 'integrations.setup.collector.sophos'
const IMG = '/integrations/guides/collector/sophos'

const FIELDS: CloudConfigField[] = [
  { key: 'sophos_client_id',  labelKey: `${ROOT}.fields.clientId.label`,  placeholder: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx' },
  { key: 'sophos_x_api_key',  labelKey: `${ROOT}.fields.apiKey.label`,    secret: true, type: 'password' },
]

const STEPS = [
  { key: 'apiCredentials', img: `${IMG}/sophos-step-1.png` },
  { key: 'addCredential',  img: `${IMG}/sophos-step-2.png` },
  { key: 'copyIds',        img: `${IMG}/sophos-step-3.png` },
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

export function SophosGuide({ module: integration }: { module: Integration }) {
  const { t } = useTranslation()
  const { configGroups: tenantsQuery, deleteConfigGroup } = useIntegrations()

  const moduleName = integration.moduleName ?? 'SOPHOS'
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
            title="Sophos Central"
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

      {/* ── API credential creation steps ────────────────────────── */}
      <SectionLabel title={t(`${ROOT}.apiSection`)} />

      {STEPS.map((s, idx) => (
        <Section key={s.key} title={t(`${ROOT}.steps.${s.key}.title`)} step={idx + 1}>
          <p className="text-sm text-foreground/90">
            <Trans
              i18nKey={`${ROOT}.steps.${s.key}.body`}
              components={{ hl: <strong className="font-semibold text-primary" /> }}
            />
          </p>
          {s.key === 'copyIds' && (
            <p className="mt-2 rounded-md bg-amber-500/10 border border-amber-500/30 px-3 py-2 text-[11px] text-amber-700 dark:text-amber-400">
              {t(`${ROOT}.steps.copyIds.secretNote`)}
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
