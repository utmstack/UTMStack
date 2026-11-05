import type { CloudGuideSection, CloudConfigField } from './cloudGuideBuilder'

const ROOT = 'integrations.setup.cloud.aws'
const IMG = '/integrations/guides/aws'

export const AWS_SECTIONS: CloudGuideSection[] = [
  {
    id: 'prerequisite',
    titleKey: `${ROOT}.sections.prerequisite.title`,
    bodyKey: `${ROOT}.sections.prerequisite.body`,
  },
  {
    id: 'create-trail',
    titleKey: `${ROOT}.sections.createTrail.title`,
    bodyKey: `${ROOT}.sections.createTrail.body`,
    image: `${IMG}/welcome-page.png`,
  },
  {
    id: 'trail-name',
    titleKey: `${ROOT}.sections.trailName.title`,
    bodyKey: `${ROOT}.sections.trailName.body`,
    image: `${IMG}/trail-name.png`,
  },
  {
    id: 's3-buckets',
    titleKey: `${ROOT}.sections.s3Buckets.title`,
    bodyKey: `${ROOT}.sections.s3Buckets.body`,
    image: `${IMG}/s3-buckets.png`,
  },
  {
    id: 'bucket-name',
    titleKey: `${ROOT}.sections.bucketName.title`,
    bodyKey: `${ROOT}.sections.bucketName.body`,
    image: `${IMG}/bucket-name.png`,
  },
  {
    id: 'edit-trail',
    titleKey: `${ROOT}.sections.editTrail.title`,
    bodyKey: `${ROOT}.sections.editTrail.body`,
    image: `${IMG}/edit-trail.png`,
  },
  {
    id: 'cloudwatch-config',
    titleKey: `${ROOT}.sections.cloudwatchConfig.title`,
    bodyKey: `${ROOT}.sections.cloudwatchConfig.body`,
    image: `${IMG}/cloud-watch-config.png`,
  },
  {
    id: 'log-group',
    titleKey: `${ROOT}.sections.logGroup.title`,
    bodyKey: `${ROOT}.sections.logGroup.body`,
    image: `${IMG}/trail-group.png`,
  },
  {
    id: 'permissions',
    titleKey: `${ROOT}.sections.permissions.title`,
    bodyKey: `${ROOT}.sections.permissions.body`,
    image: `${IMG}/trail-permissions.png`,
  },
]

export const AWS_FIELDS: CloudConfigField[] = [
  {
    key: 'aws_access_key_id',
    labelKey: `${ROOT}.fields.accessKeyId.label`,
    placeholder: 'AKIA...',
  },
  {
    key: 'aws_secret_access_key',
    labelKey: `${ROOT}.fields.secretAccessKey.label`,
    secret: true,
    type: 'password',
  },
  {
    key: 'aws_default_region',
    labelKey: `${ROOT}.fields.region.label`,
    placeholder: 'us-east-1',
  },
  {
    key: 'aws_log_group_name',
    labelKey: `${ROOT}.fields.logGroupName.label`,
    placeholder: 'utmstack',
  },
]
