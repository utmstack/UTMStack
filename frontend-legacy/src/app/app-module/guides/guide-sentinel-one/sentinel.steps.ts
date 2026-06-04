import {COLLECTOR_MESSAGE, Step} from '../shared/step';

export const SENTINELSTEPS: Step[] = [
  {id: '1', name: 'Click on the "Settings" menu in the SentinelOne sidebar.',
    content: {
      id: 'stepContent1'
    }
   },
  {id: '2', name: 'Go to "Integrations" -> "Syslog", and configure SentinelOne Endpoint ' +
      'Security to send logs to a UTMStack agent to ports:',
    content: {
      id: 'stepContent2'
    }
   },
  {id: '3', name: COLLECTOR_MESSAGE,
    content: {
      id: 'stepContent3'
    }
  },
  {id: '4', name: 'Click on the "Test" button to check connection with UTMStack, and save changes'},
  {id: '5', name: 'Click on the button shown below, to activate the UTMStack features related to this integration',
    content: {
      id: 'stepContent5'
    }
  },
];
