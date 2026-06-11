import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.awsSqlServer'

registerCollector({
  getName: () => 'AWS_SQL_SERVER',
  matches: (n) => n === 'aws_sql_server' || n === 'aws_rds_ms_sql',
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
      id: 'modify',
      titleKey: `${ROOT}.sections.modify.title`,
      bodyKey: `${ROOT}.sections.modify.body`,
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
  render: (m) => <CollectorEndpointInfo module={m} port="cloudwatch (pull)" sourceType="aws_sql_server" />,
})
