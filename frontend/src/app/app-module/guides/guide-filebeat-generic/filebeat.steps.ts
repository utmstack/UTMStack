import {Step} from '../shared/step';

export const FILEBEAT_STEPS: Step[] = [
  {id: '1', name: 'Enable Filebeat module: <br>' +
                   'To enable the AGENT_NAME module where you have the UTMStack agent installed, ' +
                   'follow the instructions below based on your operating system.',
    content: {
      id: 'stepContent1'
    }
  },
  {id: '2', name: `Configure Filebeat module:`,
    content: {
      id: 'stepContent2'
    }
  },
  {id: '3', name: 'Click on the button shown below, to activate the UTMStack features related to this integration',
    content: {
      id: 'stepContent3'
    }
  }
];
