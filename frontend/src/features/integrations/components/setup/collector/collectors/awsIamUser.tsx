import { registerCollector } from '../registry'
import { CollectorEndpointInfo } from '../CollectorEndpointInfo'

const ROOT = 'integrations.setup.collector.awsIamUser'
const IMG = '/integrations/guides/collector/aws-iam'

registerCollector({
  getName: () => 'AWS_IAM_USER',
  sections: [
    {
      id: 'iam-console',
      titleKey: `${ROOT}.sections.iamConsole.title`,
      bodyKey: `${ROOT}.sections.iamConsole.body`,
      image: `${IMG}/1.png`,
    },
    {
      id: 'create-user',
      titleKey: `${ROOT}.sections.createUser.title`,
      bodyKey: `${ROOT}.sections.createUser.body`,
      image: `${IMG}/2.png`,
    },
    {
      id: 'programmatic-access',
      titleKey: `${ROOT}.sections.programmaticAccess.title`,
      bodyKey: `${ROOT}.sections.programmaticAccess.body`,
      image: `${IMG}/3.png`,
    },
    {
      id: 'attach-policy',
      titleKey: `${ROOT}.sections.attachPolicy.title`,
      bodyKey: `${ROOT}.sections.attachPolicy.body`,
      image: `${IMG}/4.png`,
    },
    {
      id: 'review',
      titleKey: `${ROOT}.sections.review.title`,
      bodyKey: `${ROOT}.sections.review.body`,
      image: `${IMG}/5.png`,
    },
    {
      id: 'copy-keys',
      titleKey: `${ROOT}.sections.copyKeys.title`,
      bodyKey: `${ROOT}.sections.copyKeys.body`,
      image: `${IMG}/6.png`,
    },
    {
      id: 'create-log-group',
      titleKey: `${ROOT}.sections.createLogGroup.title`,
      bodyKey: `${ROOT}.sections.createLogGroup.body`,
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
  render: (m) => <CollectorEndpointInfo module={m} port="cloudwatch (pull)" sourceType="aws_iam_user" />,
})
