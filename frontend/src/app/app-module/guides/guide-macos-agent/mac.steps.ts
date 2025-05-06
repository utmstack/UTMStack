import {Step} from '../shared/step';

export const MAC_STEPS: Step[] = [
  {id: '1',
    name: 'Contact UTMStack support to obtain the <strong> `utmstack-macos-agent.pkg` </strong> file. ' +
      'This package is required to download and install the necessary dependencies.'
  },
  {
    id: '2',
    name: 'Run the <strong> `utmstack-macos-agent.pkg` file. This will download and install the required components for the UTMStack agent.'
  },
  {id: '3',
    name: 'Use the following command according to the action you wish to perform (install or uninstall):',
    content: {
      id: 'stepContent3'
    }
  },
];
