import {Step} from '../shared/step';

export const AS400STEPS: Step[] = [
  {id: '1', name: 'Install UTMStack AS400 Collector. <br>' +
                  'To install UTMStack AS400 Collector, follow the instructions below based on your operative system. <br>',
    content: {
      id: 'stepContent1'
    }
   },
  {id: '2', name: 'Add your collector configurations',
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
