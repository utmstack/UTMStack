import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.windowsFaa'
const IMG = '/integrations/guides/collector/windows-faa'

registerCollector({
  getName: () => 'WINDOWS_FAA',
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
      id: 'activate',
      titleKey: `${ROOT}.sections.activate.title`,
      bodyKey: `${ROOT}.sections.activate.body`,
    },
  ],
  render: (m) => <CollectorEndpointInfo module={m} port="agent (local)" sourceType="windows_faa" />,
})
