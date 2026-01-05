import {COLLECTOR_MESSAGE, Step} from '../shared/step';

export const CROWDSTRIKE_STEPS: Step[] = [
  {
    id: '1',
    name: 'Open the main menu. Click the hamburger icon in the top‑left corner labeled "Open menu".' +
      'Navigate to API management. Select **Support and resources** then **API Clients and Keys**.',
    content: {
      id: 'stepContent1',
      images: [{
        alt: 'Open menu hamburger top-left',
        src: '../../../../assets/img/guides/crowdstrike/1.png',
      }]
    }
  },
  {
    id: '2',
    name: 'Create a new API client. Click **Create API Client**, ' +
      'provide a descriptive client name (used to identify event sources) and select the API scopes required for Event Streams.',
    content: {
      id: 'stepContent2',
      images: [{
        alt: 'Create API client form with name and scopes',
        src: '../../../../assets/img/guides/crowdstrike/2.png',
      }]
    }
  },
  {
    id: '3',
    name: 'Generate credentials. Click **Create** to generate the client credentials and endpoint information.',
    content: {
      id: 'stepContent3',
      images: [{
        alt: 'Create button to generate API credentials',
        src: '../../../../assets/img/guides/crowdstrike/3.png',
      }]
    }
  },
  {
    id: '4',
    name: 'Record the credentials securely. Note the **Client ID**, **Client Secret** and the **Base URL** for the selected region.',
    content: {
      id: 'stepContent4',
      images: [{
        alt: 'Screen showing Client ID Client Secret and Base URL',
        src: '../../../../assets/img/guides/crowdstrike/4.png',
      }]
    }
  },
  {
    id: '5',
    name: 'Insert information in the following inputs.You can add more than one CrowdStrike configuration by clicking on Add tenant button.',
    content: {
      id: 'stepContent5'
    }
  },
  {id: '6', name: 'Click on the button shown below, to activate the UTMStack features related to this integration',
    content: {
      id: 'stepContent6'
    }
  }
];
