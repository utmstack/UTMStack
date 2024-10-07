import {COLLECTOR_MESSAGE, Step} from '../shared/step';

export const MAC_STEPS: Step[] = [
  {id: '1', name: 'Install Homebrew, using the official documentation\n' +
      '          <a class="text-primary font-weight-semibold"\n' +
      '             href="https://docs.brew.sh/Installation"\n' +
      '             target="_blank">here</a>, if you already installed go to the next step.'},
  {id: '2', name: 'Install rsyslog on MacOS:',
    content: {
      id: 'stepContent2',
      commands: ['brew install rsyslog']
    }
  },
  {id: '3', name: 'Open the file /etc/syslog.conf in an editor:\n',
    content: {
      id: 'stepContent3',
      commands: ['sudo nano /etc/syslog.conf']
    }
  },
  {id: '4', name: 'Append the following line at the end if you want to send over TPC:\n',
    content: {
      id: 'stepContent4',
      commands: ['*.* @@IP_OF_YOUR_UTMSTACK_AGENT:7015']
    }
  },
  {id: '5', name: 'Restart the syslog daemon:',
    content: {
      id: 'stepContent5',
      commands: ['sudo launchctl stop /System/Library/LaunchDaemons/com.apple.syslogd.plist',
                 'sudo launchctl start /System/Library/LaunchDaemons/com.apple.syslogd.plist']
    }
  },
  {id: '6', name: COLLECTOR_MESSAGE,
    content: {
      id: 'stepContent6'
    }
  },
  {id: '7', name: 'Click on the button shown below, to activate the UTMStack features related to this integration',
    content: {
      id: 'stepContent7'
    }
  }
];
