import { useState } from 'react'
import { useTranslation, Trans } from 'react-i18next'
import { Cloud, Database, Loader2, ShieldCheck } from 'lucide-react'
import { Section } from '@/features/integrations/components/ui/Section'
import { FlowNode, FlowEdge } from '../collector/ForwarderGuide'
import { CloudTenantList } from './CloudTenantList'
import { CloudTenantForm } from './CloudTenantForm'
import { useIntegrations } from '@/features/integrations/hooks/useIntegrations'
import { AWS_FIELDS } from './builders/aws'
import type { Integration, TenantResponse } from '@/features/integrations/types'

const ROOT = 'integrations.setup.cloud.aws'
const IAM = '/integrations/guides/aws/iam'
const CW = '/integrations/guides/aws/cloudwatch'
const CT = `${CW}/cloudtrail`

const IAM_STEPS: Array<{ key: string; img: string }> = [
  { key: 'dashboard',    img: `${IAM}/1.png` },
  { key: 'addUser',      img: `${IAM}/2.png` },
  { key: 'attachPolicy', img: `${IAM}/4.png` },
  { key: 'createUser',   img: `${IAM}/5.png` },
  { key: 'openUser',     img: `${IAM}/6.png` },
  { key: 'securityTab',  img: `${IAM}/8.png` },
  { key: 'createKey',    img: `${IAM}/9.png` },
  { key: 'confirmKey',   img: `${IAM}/10.png` },
  { key: 'downloadKey',  img: `${IAM}/11.png` },
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
      <span className="text-[11px] font-semibold uppercase tracking-widest text-muted-foreground">
        {title}
      </span>
      <div className="h-px flex-1 bg-border" />
    </div>
  )
}

export function AWSGuide({ integration }: { integration: Integration }) {
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

      {/* ── Intro + flow diagram ─────────────────────────────────── */}
      <Section title={t(`${ROOT}.intro.title`)}>
        <p className="text-sm text-foreground/90">{t(`${ROOT}.intro.body`)}</p>
        <div className="mt-3 flex items-stretch justify-center gap-1 sm:gap-2">
          <FlowNode
            icon={<Database size={20} className="text-muted-foreground" />}
            title="AWS"
            sub={t(`${ROOT}.intro.diagram.source`)}
            tone="neutral"
          />
          <FlowEdge label={t(`${ROOT}.intro.diagram.edge1`)} />
          <FlowNode
            icon={<Cloud size={20} />}
            title="CloudWatch"
            sub={t(`${ROOT}.intro.diagram.cloudwatchSub`)}
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

      {/* ── Part 1: IAM User ─────────────────────────────────────── */}
      <SectionLabel title={t(`${ROOT}.iamSection`)} />

      {IAM_STEPS.map((s, idx) => (
        <Section key={s.key} title={t(`${ROOT}.iam.${s.key}.title`)} step={idx + 1}>
          <p className="text-sm text-foreground/90">
            <Trans
              i18nKey={`${ROOT}.iam.${s.key}.body`}
              components={{ strong: <strong className="font-semibold" /> }}
            />
          </p>
          <StepImg src={s.img} />
        </Section>
      ))}

      {/* ── Credential form ──────────────────────────────────────── */}
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
            fields={AWS_FIELDS}
            editing={editing}
            onCancel={() => setEditing(null)}
            onSaved={() => setEditing(null)}
          />
        </div>
      </Section>

      {/* ── Part 2: CloudWatch / CloudTrail ──────────────────────── */}
      <SectionLabel title={t(`${ROOT}.cloudwatchSection`)} />

      <Section title={t(`${ROOT}.cloudwatch.createGroup.title`)} step={1}>
        <p className="text-sm text-foreground/90">
          <Trans i18nKey={`${ROOT}.cloudwatch.createGroup.body`} components={{ strong: <strong className="font-semibold" /> }} />
        </p>
        <StepImg src={`${CW}/1.png`} />
      </Section>

      <Section title={t(`${ROOT}.cloudwatch.nameGroup.title`)} step={2}>
        <p className="text-sm text-foreground/90">
          <Trans i18nKey={`${ROOT}.cloudwatch.nameGroup.body`} components={{ strong: <strong className="font-semibold" />, code: <code className="rounded bg-muted px-1 font-mono text-xs" /> }} />
        </p>
        <StepImg src={`${CW}/3.png`} />
      </Section>

      <Section title={t(`${ROOT}.cloudwatch.openTrail.title`)} step={3}>
        <p className="text-sm text-foreground/90">
          <Trans i18nKey={`${ROOT}.cloudwatch.openTrail.body`} components={{ strong: <strong className="font-semibold" /> }} />
        </p>
        <StepImg src={`${CT}/2.png`} />
      </Section>

      <Section title={t(`${ROOT}.cloudwatch.configureTrail.title`)} step={4}>
        <p className="mb-2 text-sm text-foreground/90">{t(`${ROOT}.cloudwatch.configureTrail.body`)}</p>
        <ul className="space-y-1 text-sm text-foreground/80">
          {(t(`${ROOT}.cloudwatch.configureTrail.items`, { returnObjects: true }) as string[]).map((item, i) => (
            <li key={i} className="flex gap-2">
              <span className="mt-0.5 shrink-0 text-muted-foreground">·</span>
              <span dangerouslySetInnerHTML={{ __html: item }} />
            </li>
          ))}
        </ul>
        <StepImg src={`${CT}/5.png`} />
      </Section>

      <Section title={t(`${ROOT}.cloudwatch.eventTypes.title`)} step={5}>
        <p className="mb-3 text-sm text-foreground/90">{t(`${ROOT}.cloudwatch.eventTypes.body`)}</p>
        {(['management', 'data', 'network'] as const).map((kind) => (
          <div key={kind} className="mb-4">
            <p className="mb-1 text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
              {t(`${ROOT}.cloudwatch.eventTypes.${kind}.label`)}
            </p>
            <p className="mb-2 text-sm text-foreground/80">
              <Trans i18nKey={`${ROOT}.cloudwatch.eventTypes.${kind}.body`} components={{ strong: <strong className="font-semibold" />, code: <code className="rounded bg-muted px-1 font-mono text-xs" /> }} />
            </p>
            <StepImg src={kind === 'management' ? `${CT}/6.png` : kind === 'data' ? `${CT}/6.1.png` : `${CT}/8.png`} />
          </div>
        ))}
      </Section>

      <Section title={t(`${ROOT}.cloudwatch.review.title`)} step={6}>
        <p className="text-sm text-foreground/90">
          <Trans i18nKey={`${ROOT}.cloudwatch.review.body`} components={{ strong: <strong className="font-semibold" /> }} />
        </p>
        <StepImg src={`${CT}/12.png`} />
      </Section>

    </div>
  )
}
