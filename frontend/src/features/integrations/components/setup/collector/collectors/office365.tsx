import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.office365'
const IMG = '/integrations/guides/collector/office365'

registerCollector({
  getName: () => 'O365',
  matches: (n) => n === 'office365' || n === 'o365',
  sections: [
    {
      id: 'azure-portal',
      titleKey: `${ROOT}.sections.azurePortal.title`,
      bodyKey: `${ROOT}.sections.azurePortal.body`,
      image: `${IMG}/o365-portal.png`,
    },
    {
      id: 'app-registrations',
      titleKey: `${ROOT}.sections.appRegistrations.title`,
      bodyKey: `${ROOT}.sections.appRegistrations.body`,
      image: `${IMG}/o365-app-registration.png`,
    },
    {
      id: 'register-app',
      titleKey: `${ROOT}.sections.registerApp.title`,
      bodyKey: `${ROOT}.sections.registerApp.body`,
      image: `${IMG}/o365-register-app.png`,
    },
    {
      id: 'client-secret',
      titleKey: `${ROOT}.sections.clientSecret.title`,
      bodyKey: `${ROOT}.sections.clientSecret.body`,
      image: `${IMG}/o365-certificate-secret.png`,
    },
    {
      id: 'api-permissions',
      titleKey: `${ROOT}.sections.apiPermissions.title`,
      bodyKey: `${ROOT}.sections.apiPermissions.body`,
      image: `${IMG}/o365-api-permission.png`,
    },
    {
      id: 'activity-feed',
      titleKey: `${ROOT}.sections.activityFeed.title`,
      bodyKey: `${ROOT}.sections.activityFeed.body`,
      image: `${IMG}/o365-api-permission-request.png`,
    },
    {
      id: 'graph-permissions',
      titleKey: `${ROOT}.sections.graphPermissions.title`,
      bodyKey: `${ROOT}.sections.graphPermissions.body`,
      image: `${IMG}/o365-activity-feed.png`,
    },
    {
      id: 'delegated-permissions',
      titleKey: `${ROOT}.sections.delegatedPermissions.title`,
      bodyKey: `${ROOT}.sections.delegatedPermissions.body`,
      image: `${IMG}/o365-delegated-permission.png`,
    },
    {
      id: 'app-permissions',
      titleKey: `${ROOT}.sections.appPermissions.title`,
      bodyKey: `${ROOT}.sections.appPermissions.body`,
      image: `${IMG}/o365-app-permission.png`,
    },
    {
      id: 'note-ids',
      titleKey: `${ROOT}.sections.noteIds.title`,
      bodyKey: `${ROOT}.sections.noteIds.body`,
      image: `${IMG}/o365-overview-client.png`,
    },
    {
      id: 'fill-form',
      titleKey: `${ROOT}.sections.fillForm.title`,
      bodyKey: `${ROOT}.sections.fillForm.body`,
    },
    {
      id: 'activate',
      titleKey: `${ROOT}.sections.activate.title`,
      bodyKey: `${ROOT}.sections.activate.body`,
    },
  ],
  render: (m) => <CollectorEndpointInfo module={m} port="443/tcp" sourceType="o365" />,
})
