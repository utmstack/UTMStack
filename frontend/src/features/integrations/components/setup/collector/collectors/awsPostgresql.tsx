import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.awsPostgresql'

registerCollector({
  getName: () => 'AWS_POSTGRESQL',
  matches: (n) => n === 'aws_postgresql' || n === 'aws_rds_postgres',
  sections: [
    {
      id: 'open-rds',
      titleKey: `${ROOT}.sections.openRds.title`,
      bodyKey: `${ROOT}.sections.openRds.body`,
    },
    {
      id: 'pick-db',
      titleKey: `${ROOT}.sections.pickDb.title`,
      bodyKey: `${ROOT}.sections.pickDb.body`,
    },
    {
      id: 'log-exports',
      titleKey: `${ROOT}.sections.logExports.title`,
      bodyKey: `${ROOT}.sections.logExports.body`,
    },
    {
      id: 'apply',
      titleKey: `${ROOT}.sections.apply.title`,
      bodyKey: `${ROOT}.sections.apply.body`,
    },
    {
      id: 'activate',
      titleKey: `${ROOT}.sections.activate.title`,
      bodyKey: `${ROOT}.sections.activate.body`,
    },
  ],
  render: (m) => <CollectorEndpointInfo module={m} port="cloudwatch (pull)" sourceType="aws_postgresql" />,
})
