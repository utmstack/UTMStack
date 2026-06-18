import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.awsBeanstalk'

registerCollector({
  getName: () => 'AWS_BEANSTALK',
  sections: [
    {
      id: 'open-console',
      titleKey: `${ROOT}.sections.openConsole.title`,
      bodyKey: `${ROOT}.sections.openConsole.body`,
    },
    {
      id: 'environment',
      titleKey: `${ROOT}.sections.environment.title`,
      bodyKey: `${ROOT}.sections.environment.body`,
    },
    {
      id: 'configuration',
      titleKey: `${ROOT}.sections.configuration.title`,
      bodyKey: `${ROOT}.sections.configuration.body`,
    },
    {
      id: 'software',
      titleKey: `${ROOT}.sections.software.title`,
      bodyKey: `${ROOT}.sections.software.body`,
    },
    {
      id: 's3-log-storage',
      titleKey: `${ROOT}.sections.s3LogStorage.title`,
      bodyKey: `${ROOT}.sections.s3LogStorage.body`,
    },
    {
      id: 'cloudwatch-streaming',
      titleKey: `${ROOT}.sections.cloudwatchStreaming.title`,
      bodyKey: `${ROOT}.sections.cloudwatchStreaming.body`,
    },
    {
      id: 'activate',
      titleKey: `${ROOT}.sections.activate.title`,
      bodyKey: `${ROOT}.sections.activate.body`,
    },
  ],
  render: (m) => <CollectorEndpointInfo module={m} port="cloudwatch (pull)" sourceType="aws_beanstalk" />,
})
