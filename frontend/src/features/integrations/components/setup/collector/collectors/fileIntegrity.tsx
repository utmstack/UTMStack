import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.fileIntegrity'
const IMG = '/integrations/guides/collector/file-integrity'

registerCollector({
  getName: () => 'FILE_INTEGRITY',
  matches: (n) => n === 'file_classification' || n === 'file_integrity',
  sections: [
    {
      id: 'enable-audit-policy',
      titleKey: `${ROOT}.sections.enableAuditPolicy.title`,
      bodyKey: `${ROOT}.sections.enableAuditPolicy.body`,
      image: `${IMG}/audit-file-system-group-policy.png`,
    },
    {
      id: 'configure-folder-auditing',
      titleKey: `${ROOT}.sections.configureFolderAuditing.title`,
      bodyKey: `${ROOT}.sections.configureFolderAuditing.body`,
    },
    {
      id: 'verify-events',
      titleKey: `${ROOT}.sections.verifyEvents.title`,
      bodyKey: `${ROOT}.sections.verifyEvents.body`,
    },
    {
      id: 'activate',
      titleKey: `${ROOT}.sections.activate.title`,
      bodyKey: `${ROOT}.sections.activate.body`,
    },
  ],
  render: (m) => <CollectorEndpointInfo module={m} port="agent (local)" sourceType="file_integrity" />,
})
