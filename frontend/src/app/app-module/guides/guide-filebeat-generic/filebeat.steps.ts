import {Step} from '../shared/step';

export const FILEBEAT_STEPS: Step[] = [
  {id: '1', name: 'Enable Filebeat module: <br>' +
                   'Follow the instructions below, based on your operating system, ' +
                   'to enable the AGENT_NAME module where the Vortex agent is installed.',
    content: {
      id: 'stepContent1'
    }
  },
  {id: '2', name: `Configure Filebeat module:`,
    content: {
      id: 'stepContent2'
    }
  },
  {id: '3', name: 'Click on the button shown below, to activate the Vortex features related to this integration',
    content: {
      id: 'stepContent3'
    }
  }
];
