import { useState } from 'react'
import { useTranslation, Trans } from 'react-i18next'
import { Loader2, Shield, ShieldCheck, Wifi } from 'lucide-react'
import { Section } from '@/features/integrations/components/ui/Section'
import { FlowNode, FlowEdge } from '../ForwarderGuide'
import { CloudTenantList } from '@/features/integrations/components/setup/cloud/CloudTenantList'
import { CloudTenantForm } from '@/features/integrations/components/setup/cloud/CloudTenantForm'
import { useIntegrations } from '@/features/integrations/hooks/useIntegrations'
import type { CloudConfigField } from '@/features/integrations/components/setup/cloud/builders/cloudGuideBuilder'
import type { Integration, ConfigGroupResponse } from '@/features/integrations/types'

const ROOT = 'integrations.setup.collector.bitdefender'
const IMG = '/integrations/guides/collector/bitdefender'

const WHITELIST_IPS = [
  '34.159.83.241',
  '34.159.47.15',
  '34.159.150.228',
  '34.85.152.87',
  '34.85.155.173',
]

const FIELDS: CloudConfigField[] = [
  { key: 'accessUrl',     labelKey: `${ROOT}.fields.accessUrl.label`,     placeholder: 'https://cloud.gravityzone.bitdefender.com' },
  { key: 'connectionKey', labelKey: `${ROOT}.fields.connectionKey.label`, secret: true, type: 'password' },
  { key: 'utmPublicIP',   labelKey: `${ROOT}.fields.utmPublicIP.label`,   placeholder: '1.2.3.4' },
  { key: 'companyIds',    labelKey: `${ROOT}.fields.companyIds.label`,     placeholder: 'id1,id2,id3' },
]

const API_STEPS = [
  { key: 'myAccount',   img: `${IMG}/step2.png` },
  { key: 'accessUrl',   img: `${IMG}/step3.png` },
  { key: 'createKey',   img: `${IMG}/step4.png` },
]

const COMPANY_STEPS = [
  { key: 'myCompany',  img: `${IMG}/bdgz4.png` },
  { key: 'licensing',  img: `${IMG}/bdgz5.png` },
]

function StepImg({ src }: { src: string }) {
  return (
    <img
      src={src}
      alt=""
      className="mt-3 max-w-full rounded-md border border-border"
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

export function BitdefenderGuide({ module: integration }: { module: Integration }) {
  const { t } = useTranslation()
  const { configGroups: tenantsQuery, deleteConfigGroup } = useIntegrations()

  const moduleName = integration.moduleName ?? 'BITDEFENDER'
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
            icon={<Shield size={20} className="text-muted-foreground" />}
            title="GravityZone"
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

      {/* ── IP whitelist warning ─────────────────────────────────── */}
      <Section title={t(`${ROOT}.whitelist.title`)} variant="warning">
        <p className="mb-2 text-sm text-foreground/90">{t(`${ROOT}.whitelist.body`)}</p>
        <ul className="space-y-1">
          {WHITELIST_IPS.map((ip) => (
            <li key={ip} className="flex items-center gap-2 text-sm">
              <Wifi size={12} className="shrink-0 text-amber-500" />
              <code className="font-mono text-xs">{ip}</code>
            </li>
          ))}
        </ul>
        <p className="mt-2 text-[11px] text-muted-foreground">{t(`${ROOT}.whitelist.note`)}</p>
      </Section>

      {/* ── Part 1: API credentials ──────────────────────────────── */}
      <SectionLabel title={t(`${ROOT}.apiSection`)} />

      {API_STEPS.map((s, idx) => (
        <Section key={s.key} title={t(`${ROOT}.api.${s.key}.title`)} step={idx + 1}>
          <p className="text-sm text-foreground/90">
            <Trans i18nKey={`${ROOT}.api.${s.key}.body`} components={{ strong: <strong className="font-semibold" /> }} />
          </p>
          <StepImg src={s.img} />
        </Section>
      ))}

      {/* ── Part 2: Company IDs ──────────────────────────────────── */}
      <SectionLabel title={t(`${ROOT}.companySection`)} />

      {COMPANY_STEPS.map((s, idx) => (
        <Section key={s.key} title={t(`${ROOT}.company.${s.key}.title`)} step={idx + 1}>
          <p className="text-sm text-foreground/90">
            <Trans i18nKey={`${ROOT}.company.${s.key}.body`} components={{ strong: <strong className="font-semibold" /> }} />
          </p>
          <StepImg src={s.img} />
        </Section>
      ))}

      {/* ── Configuration form ───────────────────────────────────── */}
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
