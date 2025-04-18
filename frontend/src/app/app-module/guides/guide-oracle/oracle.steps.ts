import {Step} from '../shared/step';

export const OracleSteps: Step[] = [
  {id: '1', name: 'Run the following command to create the configuration file',
    content: {
      id: 'stepContent1',
      commands: ['sudo touch /etc/rsyslog.d/oracle-utmstack.conf']
    }
   },
  {id: '2', name: ' Open file “/etc/rsyslog.d/oracle-utmstack.conf” with your preferred editor and add the following configuration:',
    content: {
      id: 'stepContent2',
      commands: ['input(type="imfile"<br>\n' +
      '&nbsp;&nbsp;&nbsp;&nbsp;File="BASEDIR/admin/*/adump/*.aud"<br>\n' +
      '&nbsp;&nbsp;&nbsp;&nbsp;Tag="oracle-audit"<br>\n' +
      '&nbsp;&nbsp;&nbsp;&nbsp;Severity="info"<br>\n' +
      '&nbsp;&nbsp;&nbsp;&nbsp;Facility="local5")<br><br>\n' +
      '\n' +
      'input(type="imfile"<br>\n' +
      '&nbsp;&nbsp;&nbsp;&nbsp;File="BASEDIR/diag/tnslsnr/*/listener/trace/listener.log"<br>\n' +
      '&nbsp;&nbsp;&nbsp;&nbsp;Tag="oracle-listener"<br>\n' +
      '&nbsp;&nbsp;&nbsp;&nbsp;Severity="info"<br>\n' +
      '&nbsp;&nbsp;&nbsp;&nbsp;Facility="local5")<br><br>\n' +
      '\n' +
      'input(type="imfile"<br>\n' +
      '&nbsp;&nbsp;&nbsp;&nbsp;File="BASEDIR/diag/rdbms/*/*/trace/alert_*.log"<br>\n' +
      '&nbsp;&nbsp;&nbsp;&nbsp;Tag="oracle-alerts"<br>\n' +
      '&nbsp;&nbsp;&nbsp;&nbsp;Severity="info"<br>\n' +
      '&nbsp;&nbsp;&nbsp;&nbsp;Facility="local5")<br><br>\n' +
      '\n' +
      'local5.* @DESTINATION:7014\n']
    }
  },
  {id: '3', name: 'Replace BASEDIR with the actual Oracle DB working/base path (e.g., /u01/app/oracle)'},
  {id: '4', name: 'Replace DESTINATION with the UTMStack Agent’s IP address or Fully Qualified Domain Name (FQDN)',
    content: {
      id: 'stepContent4',
      commands: ['local5.* @@DESTINATION:7014']
    }
  },
  {id: '5', name: 'Restart rsyslog by running the following command',
    content: {
      id: 'stepContent5',
      commands: ['sudo systemctl restart rsyslog']
    }
  },
  {id: '6', name: 'Validate the configuration by sending test logs by running the following commands',
    content: {
      id: 'stepContent5',
      commands: ['logger -p local5.info "Test Oracle Audit log message"<br>\n' +
      'logger -p local5.info "Test Oracle Listener log message"<br>\n' +
      'logger -p local5.info "Test Oracle Alert log message"\n']
    }
  },
  {id: '7', name: 'Verify that UTMStack received these messages using the Logs Explorer and searching for the test message above.'}
];
