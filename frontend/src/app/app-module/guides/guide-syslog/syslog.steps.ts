import {Step} from '../shared/step';

export const SYSLOGSTEPS: Step[] = [
  {id: '1', name: 'Configure your device to send logs to a UTMStack agent on ports:',
    content: {
      id: 'stepContent1'
    }
   },
  {id: '2', name: 'Enable log collector.<br>' +
                   'To enable the log collector where you have the UTMStack agent installed, ' +
                    'follow the instructions below based on your operating system and preferred protocol.\n',
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
