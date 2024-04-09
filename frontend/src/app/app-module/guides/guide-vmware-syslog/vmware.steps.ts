import {Step} from '../shared/step';

export const VMWARE_STEPS: Step[] = [
  {id: '1', name: 'Log in to your VMware vSphere Client.'},
  {id: '2', name: 'Select the host that manages your VMware inventory'},
  {id: '3', name: 'Click on the Configuration tab.'},
  {id: '4', name: 'From the Software panel, click on Advanced Settings'},
  {id: '5', name: 'In the navigation menu, click on Syslog.'},
  {id: '6', name: 'Configure values for the following parameters:', content: { id: 'stepContent6'}},
  {id: '7', name: 'Click OK to save the configuration'},
  {id: '8', name: 'Log in to your VMware ESXi Server.'},
  {id: '9', name: 'Configure Local and Remote Logging: open a ESXi Shell console session where the esxcli command is available, ' +
                    'such as the vCLI or on the ESXi host directly.'},
  {id: '10', name: 'Display the existing five configuration options on the host by running this command:',
    content: {
      id: 'stepContent10',
      commands: ['esxcli system syslog config get']
    }
   },
  {id: '11', name: 'Set new host configuration, specifying options to change, by running a command:',
    content: {
      id: 'stepContent11',
      commands: ['esxcli system syslog config set --loghost=\'tcp://your_utmstack_agent_ip:7002’',
                 'esxcli system syslog config set --logdir=/scratch/log --loghost=your_utmstack_agent_ip --logdir-unique=true']
    }
  },
  {id: '12', name: 'After making configuration changes, load the new configuration by running this command:',
    content: {
      id: 'stepContent12',
      commands: ['esxcli system syslog reload']
    }
  },
  {id: '13', name: 'Configuring ESXi Firewall Exception using the esxcli command/syslog port',
    content: {
      id: 'stepContent13',
      commands: ['esxcli network firewall ruleset set --ruleset-id=syslog --enabled=true',
                 'esxcli network firewall refresh']
    }
  },
  {id: '14', name: 'Run this command to test if the port is reachable from the ESXi host:',
    content: {
      id: 'stepContent14',
      commands: ['nc -z your_utmstack_agent_ip 7002']
    }
  },
  {id: '15', name: 'Enable log collector.<br>' +
                   'To enable the log collector where you have the UTMStack agent installed, ' +
                    'follow the instructions below based on your operating system and preferred protocol.\n',
    content: {
      id: 'stepContent15'
    }
  },
  {id: '16', name: 'Click on the button shown below, to activate the UTMStack features related to this integration',
    content: {
      id: 'stepContent16'
    }
  },
];
