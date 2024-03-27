import {Step} from '../shared/step';

export const ESET_STEPS: Step[] = [
  {id: '1', name: 'Click on the "More" menu in the ESET Protect sidebar.',
    content: {
      id: 'stepContent1',
      images: [{
        alt: 'IAM configuration',
        src: '../../../../assets/img/guides/eset/main_page.png',
      }]
    }
  },
  {id: '2', name: 'Click on the "Server configuration" submenu in the ESET Protect sidebar.',
    content: {
      id: 'stepContent2',
      images: [{
        alt: 'Programmatic access',
        src: '../../../../assets/img/guides/eset/more_settings.png',
      }]
    }
  },
  {id: '3', name: 'Click on "Advanced settings" to deploy the list of configurations.',
    content: {
      id: 'stepContent3',
      images: [{
        alt: 'CloudWatchReadOnlyAccess',
        src: '../../../../assets/img/guides/eset/server_config.png',
      }]
    }
  },
  {id: '4', name: 'Scroll down to "Syslog server", enable the “Export logs to Syslog” option and configure ESET Protect to send logs to a UTMStack agent to ports:',
    content: {
      id: 'stepContent4'
    }
  },
  {id: '5', name: 'Enable log collector.<br>' +
                   'To enable the log collector where you have the UTMStack agent installed, ' +
                    'follow the instructions below based on your operating system and preferred protocol.\n',
    content: {
      id: 'stepContent5'
    }
  },
  {id: '6', name: 'Click on the button shown below, to activate the UTMStack features related to this integration',
    content: {
      id: 'stepContent6'
    }
  },
];
