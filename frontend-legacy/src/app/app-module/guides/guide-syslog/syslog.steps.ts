import {COLLECTOR_MESSAGE, Step} from '../shared/step';

export const SYSLOGSTEPS: Step[] = [
  {id: '1', name: 'Configure your device to send logs to a UTMStack agent on ports:',
    content: {
      id: 'stepContent1'
    }
   },
  {id: '2', name: COLLECTOR_MESSAGE,
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
