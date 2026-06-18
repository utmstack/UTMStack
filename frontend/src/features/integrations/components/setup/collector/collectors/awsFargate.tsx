import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.awsFargate'

registerCollector({
  getName: () => 'AWS_FARGATE',
  matches: (n) => n === 'aws_fargate' || n === 'aws_ecs_fargate',
  sections: [
    {
      id: 'task-definition',
      titleKey: `${ROOT}.sections.taskDefinition.title`,
      bodyKey: `${ROOT}.sections.taskDefinition.body`,
    },
    {
      id: 'cloudwatch-logs',
      titleKey: `${ROOT}.sections.cloudwatchLogs.title`,
      bodyKey: `${ROOT}.sections.cloudwatchLogs.body`,
    },
    {
      id: 'iam-prerequisite',
      titleKey: `${ROOT}.sections.iamPrerequisite.title`,
      bodyKey: `${ROOT}.sections.iamPrerequisite.body`,
    },
    {
      id: 'activate',
      titleKey: `${ROOT}.sections.activate.title`,
      bodyKey: `${ROOT}.sections.activate.body`,
    },
  ],
  render: (m) => <CollectorEndpointInfo module={m} port="cloudwatch (pull)" sourceType="aws_fargate" />,
})
