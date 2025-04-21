import {Step} from '../shared/step';

export const MAC_STEPS: Step[] = [
  {id: '1',
    name: 'Reach out to support to request the installation dependencies for macOS. ' +
      'These are required to proceed with the installation or uninstallation process.',
  },
  {id: '2',
    name: 'Use the following command according to the action you wish to perform (install or uninstall):',
    content: {
      id: 'stepContent2'
    }
  },
];
