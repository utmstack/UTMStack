import {Step} from '../shared/step';

export const SOPHOS_STEPS: Step[] = [
  {id: '1', name: 'You require a Client ID and Client Secret to access event data via the API. ' +
                  'In Sophos Central Admin, go to <strong> Global Settings > API Credentials Management </strong>. <br>',
  },
  {id: '2', name: 'To create a new credential, click Add Credential from the top-right corner of the screen'},
  {id: '3', name: 'Enter a name and description for the credential, then select the role you want to assign and click Add.'},
  {id: '4', name: 'Click Show Client Secret to view the Client ID and Client Secret, then click Copy to store them securely. <br>' +
                  '<div class="w-100 alert alert-info alert-styled-right mb-3 alert-dismissible">' +
                    'The Client Secret is only visible once. Ensure you copy and save it securely</div>',
    content: {
      id: 'stepContent4',
      images: [{
        alt: 'Client Secrets',
        src: '../../../../assets/img/guides/sophos/sophos-step-4.png',
      }]
    }
  },
  {id: '5', name: 'Insert information in the following inputs.You can add more than one Sophos configuration ' +
                  'by clicking on Add tenant button.',
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
