import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.awsLambda'

registerCollector({
  getName: () => 'AWS_LAMBDA',
  sections: [
    {
      id: 'iam-prerequisite',
      titleKey: `${ROOT}.sections.iamPrerequisite.title`,
      bodyKey: `${ROOT}.sections.iamPrerequisite.body`,
    },
    {
      id: 'cloudwatch-stream',
      titleKey: `${ROOT}.sections.cloudwatchStream.title`,
      bodyKey: `${ROOT}.sections.cloudwatchStream.body`,
    },
    {
      id: 'activate',
      titleKey: `${ROOT}.sections.activate.title`,
      bodyKey: `${ROOT}.sections.activate.body`,
    },
  ],
  render: (m) => <CollectorEndpointInfo module={m} port="cloudwatch (pull)" sourceType="aws_lambda" />,
})
