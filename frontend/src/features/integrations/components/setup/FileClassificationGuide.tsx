import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ChevronDown } from 'lucide-react'
import { AgentSetup } from '@/features/integrations/components/setup/AgentSetup'
import { Section } from '@/features/integrations/components/ui/Section'
import type { Integration } from '@/features/integrations/types'

const ROOT = 'integrations.fileClassificationGuide'
const GUIDE_IMAGES = '/integrations/guides/collector/windows-faa'
const POLICY_IMAGE = `${GUIDE_IMAGES}/audit-file-system-group-policy.png`
const WINDOWS_AGENT: Integration = {
  id: 'file-classification-windows-agent-guide',
  name: 'Windows Agent',
  moduleName: 'WINDOWS_AGENT',
  dataType: 'wineventlog',
  kind: 'agents & syslog',
  status: 'available',
  description: '',
  category: 'operating-system',
  logo: '/integrations/windows.png',
}

function GuideImage({ src, alt }: { src: string; alt: string }) {
  return (
    <img
      src={src}
      alt={alt}
      loading="lazy"
      className="mx-auto mt-4 max-w-full rounded-md border border-border"
    />
  )
}

function InstructionList({
  text,
  images = [],
}: {
  text: string
  images?: { afterStep: number; src: string; alt: string }[]
}) {
  return (
    <ol className="list-decimal space-y-2 pl-5 text-sm text-foreground/90">
      {text.split('\n').map((instruction, index) => (
        <li key={`${index}-${instruction}`}>
          {instruction}
          {images
            .filter((image) => image.afterStep === index + 1)
            .map((image) => <GuideImage key={image.src} src={image.src} alt={image.alt} />)}
        </li>
      ))}
    </ol>
  )
}

export function FileClassificationGuide() {
  const { t } = useTranslation()
  const [showAgentInstall, setShowAgentInstall] = useState(false)
  const fields = t(`${ROOT}.fields`).split('\n')

  return (
    <div className="space-y-4">
      <Section title={t(`${ROOT}.overview.title`)}>
        <p className="text-sm text-foreground/90">{t(`${ROOT}.overview.body`)}</p>
      </Section>

      <Section title={t(`${ROOT}.agentInstall.title`)}>
        <p className="mb-3 text-sm text-foreground/90">{t(`${ROOT}.agentInstall.body`)}</p>
        <button
          type="button"
          onClick={() => setShowAgentInstall((open) => !open)}
          aria-expanded={showAgentInstall}
          className="flex w-full items-center justify-between rounded-md border border-border px-3 py-2 text-left text-sm font-medium hover:bg-muted/40"
        >
          <span>{t(`${ROOT}.agentInstall.button`)}</span>
          <ChevronDown size={14} className={showAgentInstall ? 'rotate-180' : ''} />
        </button>
        {showAgentInstall && (
          <div className="mt-3">
            <AgentSetup integration={WINDOWS_AGENT} />
          </div>
        )}
      </Section>

      <Section title={t(`${ROOT}.fieldsTitle`)}>
        <ul className="grid gap-x-6 gap-y-1 text-sm text-foreground/90 sm:grid-cols-2">
          {fields.map((field) => <li key={field}>{field}</li>)}
        </ul>
      </Section>

      <Section title={t(`${ROOT}.server.policyTitle`)} step={1}>
        <InstructionList
          text={t(`${ROOT}.server.policySteps`)}
          images={[
            {
              afterStep: 2,
              src: POLICY_IMAGE,
              alt: t(`${ROOT}.policyImageAlt`),
            },
            {
              afterStep: 3,
              src: `${GUIDE_IMAGES}/audit-file-system-policy-settings.png`,
              alt: t(`${ROOT}.images.serverPolicySettings`),
            },
          ]}
        />
      </Section>

      <Section title={t(`${ROOT}.server.folderTitle`)} step={2}>
        <p className="mb-3 text-sm text-foreground/90">{t(`${ROOT}.server.folderIntro`)}</p>
        <InstructionList
          text={t(`${ROOT}.server.folderSteps`)}
          images={[
            {
              afterStep: 1,
              src: `${GUIDE_IMAGES}/server-folder-properties.png`,
              alt: t(`${ROOT}.images.serverFolderProperties`),
            },
            {
              afterStep: 2,
              src: `${GUIDE_IMAGES}/server-folder-advanced-security.png`,
              alt: t(`${ROOT}.images.serverFolderAdvancedSecurity`),
            },
            {
              afterStep: 3,
              src: `${GUIDE_IMAGES}/server-folder-auditing-entry.png`,
              alt: t(`${ROOT}.images.serverFolderAuditingEntry`),
            },
          ]}
        />
      </Section>

      <Section title={t(`${ROOT}.desktop.policyTitle`)} step={1}>
        <InstructionList
          text={t(`${ROOT}.desktop.policySteps`)}
          images={[
            {
              afterStep: 3,
              src: `${GUIDE_IMAGES}/audit-object-access-policy-list.png`,
              alt: t(`${ROOT}.images.desktopPolicyList`),
            },
            {
              afterStep: 5,
              src: `${GUIDE_IMAGES}/audit-object-access-policy-settings.png`,
              alt: t(`${ROOT}.images.desktopPolicySettings`),
            },
          ]}
        />
      </Section>

      <Section title={t(`${ROOT}.desktop.folderTitle`)} step={2}>
        <p className="mb-3 text-sm text-foreground/90">{t(`${ROOT}.desktop.folderIntro`)}</p>
        <InstructionList
          text={t(`${ROOT}.desktop.folderSteps`)}
          images={[
            {
              afterStep: 1,
              src: `${GUIDE_IMAGES}/desktop-folder-security-properties.png`,
              alt: t(`${ROOT}.images.desktopFolderProperties`),
            },
            {
              afterStep: 2,
              src: `${GUIDE_IMAGES}/desktop-folder-advanced-security.png`,
              alt: t(`${ROOT}.images.desktopFolderAdvancedSecurity`),
            },
            {
              afterStep: 4,
              src: `${GUIDE_IMAGES}/desktop-folder-select-principal.png`,
              alt: t(`${ROOT}.images.desktopFolderSelectPrincipal`),
            },
            {
              afterStep: 6,
              src: `${GUIDE_IMAGES}/desktop-folder-auditing-permissions.png`,
              alt: t(`${ROOT}.images.desktopFolderAuditingPermissions`),
            },
          ]}
        />
      </Section>

      <Section title={t(`${ROOT}.events.title`)} step={3}>
        <p className="text-sm text-foreground/90">{t(`${ROOT}.events.body`)}</p>
        <GuideImage
          src={`${GUIDE_IMAGES}/event-viewer-file-access-log.png`}
          alt={t(`${ROOT}.images.events`)}
        />
      </Section>
    </div>
  )
}
