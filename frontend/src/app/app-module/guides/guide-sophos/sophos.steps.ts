import {Step} from '../shared/step';

export const SOPHOS_STEPS: Step[] = [
  {id: '1', name: 'Navigate to <strong> General Settings -> API Credentials Management </strong> in Sophos Central Admin. <br>',
    content: {
      id: 'stepContent1',
      images: [{
        alt: 'Api Credentials',
        src: '../../../../assets/img/guides/sophos/sophos-step-1.png',
      }]
    }
  },
  {id: '2', name: 'Create a New Credential:' +
                  '<ul class="mt-2 ml-4">\n' +
                  '            <li>Click <strong>Add Credential</strong> (usually found at the top-right).</li>\n' +
                  '            <li>Provide <strong>Name</strong> and <strong>Description</strong>.</li>\n' +
                  '            <li>Select the appropriate <strong>Role</strong>.</li>\n' +
                  '            <li>Click <strong>Add</strong>.</li>\n' +
                  '        </ul>',
    content: {
      id: 'stepContent2',
      images: [{
        alt: 'New Credentials',
        src: '../../../../assets/img/guides/sophos/sophos-step-2.png',
      }]
    }
  },
  {id: '3', name: 'Copy the Client ID and Client Secret and store them securely. <br>' +
                  '<div class="w-100 alert alert-info alert-styled-right mt-1 mb-2 alert-dismissible">' +
                    'The Client Secret is visible only once; ensure you save it somewhere safe.</div>',
    content: {
      id: 'stepContent3',
      images: [{
        alt: 'Client Secrets',
        src: '../../../../assets/img/guides/sophos/sophos-step-3.png',
      }]
    }
  },
  {id: '4', name: 'Insert information in the following inputs.You can add more than one Sophos configuration ' +
                  'by clicking on <strong> Add tenant </strong> button.',
    content: {
      id: 'stepContent4'
    }
  },
  {id: '5', name: 'Click on the button shown below, to activate the UTMStack features related to this integration',
    content: {
      id: 'stepContent5'
    }
  }
];
