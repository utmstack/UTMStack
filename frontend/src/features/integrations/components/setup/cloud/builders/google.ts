import type { CloudGuideSection, CloudConfigField } from './cloudGuideBuilder'

const ROOT = 'integrations.setup.cloud.google'
const IMG = '/integrations/guides/google'

export const GOOGLE_SECTIONS: CloudGuideSection[] = [
  {
    id: 'create-topic',
    titleKey: `${ROOT}.sections.createTopic.title`,
    bodyKey: `${ROOT}.sections.createTopic.body`,
    image: `${IMG}/console-topic.png`,
  },
  {
    id: 'create-subscription',
    titleKey: `${ROOT}.sections.createSubscription.title`,
    bodyKey: `${ROOT}.sections.createSubscription.body`,
    image: `${IMG}/console-newsub.png`,
  },
  {
    id: 'configure-subscription',
    titleKey: `${ROOT}.sections.configureSubscription.title`,
    bodyKey: `${ROOT}.sections.configureSubscription.body`,
    image: `${IMG}/console-editsub.png`,
  },
  {
    id: 'logs-router',
    titleKey: `${ROOT}.sections.logsRouter.title`,
    bodyKey: `${ROOT}.sections.logsRouter.body`,
    image: `${IMG}/log-router.png`,
  },
  {
    id: 'sink-destination',
    titleKey: `${ROOT}.sections.sinkDestination.title`,
    bodyKey: `${ROOT}.sections.sinkDestination.body`,
    image: `${IMG}/sink-destination.png`,
  },
  {
    id: 'service-account',
    titleKey: `${ROOT}.sections.serviceAccount.title`,
    bodyKey: `${ROOT}.sections.serviceAccount.body`,
    image: `${IMG}/create-service-account.png`,
  },
  {
    id: 'service-account-key',
    titleKey: `${ROOT}.sections.serviceAccountKey.title`,
    bodyKey: `${ROOT}.sections.serviceAccountKey.body`,
    image: `${IMG}/newkey.png`,
  },
  {
    id: 'download-key',
    titleKey: `${ROOT}.sections.downloadKey.title`,
    bodyKey: `${ROOT}.sections.downloadKey.body`,
    image: `${IMG}/downloadkey.png`,
  },
]

export const GOOGLE_FIELDS: CloudConfigField[] = [
  {
    key: 'projectId',
    labelKey: `${ROOT}.fields.projectId.label`,
    placeholder: 'my-gcp-project',
  },
  {
    key: 'topic',
    labelKey: `${ROOT}.fields.topic.label`,
    placeholder: 'utmstack-topic',
  },
  {
    key: 'subscription',
    labelKey: `${ROOT}.fields.subscription.label`,
    placeholder: 'utmstack-subscription',
  },
  {
    key: 'jsonKey',
    labelKey: `${ROOT}.fields.jsonKey.label`,
    type: 'file',
    accept: 'application/json,.json',
  },
]
