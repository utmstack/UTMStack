import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.ldap'
const IMG = '/integrations/guides/collector/ldap'

registerCollector({
  getName: () => 'AD_AUDIT',
  matches: (n) => n === 'ldap' || n === 'ad_audit',
  sections: [
    {
      id: 'open-server-manager',
      titleKey: `${ROOT}.sections.openServerManager.title`,
      bodyKey: `${ROOT}.sections.openServerManager.body`,
      image: `${IMG}/picture-1.png`,
    },
    {
      id: 'add-roles',
      titleKey: `${ROOT}.sections.addRoles.title`,
      bodyKey: `${ROOT}.sections.addRoles.body`,
      image: `${IMG}/picture-2.png`,
    },
    {
      id: 'select-cert-services',
      titleKey: `${ROOT}.sections.selectCertServices.title`,
      bodyKey: `${ROOT}.sections.selectCertServices.body`,
      image: `${IMG}/picture-3.png`,
    },
    {
      id: 'install-services',
      titleKey: `${ROOT}.sections.installServices.title`,
      bodyKey: `${ROOT}.sections.installServices.body`,
      image: `${IMG}/picture-4.png`,
    },
    {
      id: 'configure-ad-cs',
      titleKey: `${ROOT}.sections.configureAdCs.title`,
      bodyKey: `${ROOT}.sections.configureAdCs.body`,
      image: `${IMG}/picture-5.png`,
    },
    {
      id: 'pick-ca-type',
      titleKey: `${ROOT}.sections.pickCaType.title`,
      bodyKey: `${ROOT}.sections.pickCaType.body`,
      image: `${IMG}/picture-6.png`,
    },
    {
      id: 'set-validity',
      titleKey: `${ROOT}.sections.setValidity.title`,
      bodyKey: `${ROOT}.sections.setValidity.body`,
      image: `${IMG}/picture-7.png`,
    },
    {
      id: 'confirm',
      titleKey: `${ROOT}.sections.confirm.title`,
      bodyKey: `${ROOT}.sections.confirm.body`,
      image: `${IMG}/picture-8.png`,
    },
    {
      id: 'restart',
      titleKey: `${ROOT}.sections.restart.title`,
      bodyKey: `${ROOT}.sections.restart.body`,
    },
    {
      id: 'activate',
      titleKey: `${ROOT}.sections.activate.title`,
      bodyKey: `${ROOT}.sections.activate.body`,
    },
  ],
  render: (m) => <CollectorEndpointInfo module={m} port="636/tcp (LDAPS)" sourceType="ad_audit" />,
})
