import {COLLECTOR_MESSAGE, Step} from '../shared/step';

export const KASP_STEPS: Step[] = [
  {id: '1', name: 'Click on the “Console Settings” menu in the Kaspersky Security sidebar, in click in the “Integration” submenu.',
    content: {
      id: 'stepContent1',
      images: [{
        alt: 'Console Setting',
        src: '../../../../assets/img/guides/kaspersky/main_page.png',
      }]
    }
  },
  {id: '2', name: 'Click on the "Integration" -> "SIEM" in the integration section, and click on the "Settings" link.\n',
    content: {
      id: 'stepContent2',
      images: [{
        alt: 'Programmatic access',
        src: '../../../../assets/img/guides/kaspersky/integration.png',
      }]
    }
  },
  {id: '3', name: 'Configure Kaspersky Security to send logs in Splunk(CEF) format to a UTMStack agent to ports:',
    content: {
      id: 'stepContent3'
    }
  },
  {id: '4', name: COLLECTOR_MESSAGE,
    content: {
      id: 'stepContent4'
    }
  },
  {id: '5', name: 'Click on the button shown below, to activate the UTMStack features related to this integration',
    content: {
      id: 'stepContent5'
    }
  },
];
