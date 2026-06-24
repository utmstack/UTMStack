import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Cloud, ExternalLink, Loader2, Rss, ShieldCheck } from 'lucide-react'
import { Section } from '@/features/integrations/components/ui/Section'
import { FlowNode, FlowEdge } from '../collector/ForwarderGuide'
import { CloudTenantList } from './CloudTenantList'
import { CloudTenantForm } from './CloudTenantForm'
import { useIntegrations } from '@/features/integrations/hooks/useIntegrations'
import { GOOGLE_FIELDS } from './builders/google'
import type { Integration, TenantResponse } from '@/features/integrations/types'

const ROOT = 'integrations.setup.cloud.google'
const IMG = '/integrations/guides/google'

// Numbered steps condense the legacy Google Pub/Sub guide: create topic →
// subscription → Logs Router sink → service account → JSON key. Images come from
// the legacy guide assets; a couple of steps carry an external GCP console link.
interface GcpStep {
  key: string
  img?: string
  doc?: string
}

const STEPS: GcpStep[] = [
  { key: 'createTopic', img: `${IMG}/console-topic.png`, doc: 'https://console.cloud.google.com/cloudpubsub/topic/list' },
  { key: 'createSubscription', img: `${IMG}/console-newsub.png` },
  { key: 'configureSubscription', img: `${IMG}/console-editsub.png` },
  { key: 'logsRouter', img: `${IMG}/log-router.png` },
  { key: 'sinkDestination', img: `${IMG}/sink-destination.png` },
  { key: 'serviceAccount', img: `${IMG}/create-service-account.png`, doc: 'https://console.cloud.google.com/iam-admin/serviceaccounts' },
  { key: 'serviceAccountKey', img: `${IMG}/newkey.png` },
  { key: 'downloadKey', img: `${IMG}/downloadkey.png` },
]

const MAPPING_KEYS = ['projectId', 'topicId', 'subscription', 'jsonKey'] as const

export function GcpGuide({ integration }: { integration: Integration }) {
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
      {/* Intro + flow diagram. GCP is a puller running inside UTMStack: a Logs
          Router sink streams logs to a Pub/Sub topic and UTMStack pulls the
          subscription — nothing is installed in GCP. */}
      <Section title={t(`${ROOT}.intro.title`)}>
        <p className="text-sm text-foreground/90">{t(`${ROOT}.intro.body`)}</p>
        <div className="mt-3 flex items-stretch justify-center gap-1 sm:gap-2">
          <FlowNode
            icon={<Cloud size={20} className="text-muted-foreground" />}
            title="Google Cloud"
            sub={t(`${ROOT}.intro.diagram.source`)}
            tone="neutral"
          />
          <FlowEdge label={t(`${ROOT}.intro.diagram.edge1`)} />
          <FlowNode
            icon={<Rss size={20} />}
            title={t(`${ROOT}.intro.diagram.pubsub`)}
            sub={t(`${ROOT}.intro.diagram.pubsubSub`)}
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

      {/* Numbered GCP setup steps. */}
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
            fields={GOOGLE_FIELDS}
            editing={editing}
            onCancel={() => setEditing(null)}
            onSaved={() => setEditing(null)}
          />
        </div>
      </Section>
    </div>
  )
}
