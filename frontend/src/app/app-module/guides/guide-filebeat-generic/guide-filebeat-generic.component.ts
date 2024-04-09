import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-guide-filebeat-generic',
  templateUrl: './guide-filebeat-generic.component.html',
  styleUrls: ['./guide-filebeat-generic.component.css']
})
export class GuideFilebeatGenericComponent implements OnInit {
  @Input() integrationId: number;
  @Input() filebeatModule: UtmModulesEnum;
  @Input() filebeatModuleName: string;
  module = UtmModulesEnum;
  @Input() serverId: number;
  commandsActivate: FilebeatCommands[] = [
    {module: UtmModulesEnum.TRAEFIK, os: 'linux', command: 'cd /opt/utmstack-linux-agent/beats/filebeat/ && '
        + './filebeat modules enable traefik',
      path: '/opt/utmstack-linux-agent/beats/filebeat/modules.d/', restartCmd: 'sudo systemctl restart UTMStackModulesLogsCollector'},
    {module: UtmModulesEnum.TRAEFIK, os: 'windows', command: 'cd "C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\" '
        + '&& filebeat modules enable traefik', path: 'C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\modules.d\\',
      restartCmd: 'sc stop UTMStackModulesLogsCollector && timeout /t 5 && sc start UTMStackModulesLogsCollector'},

    {module: UtmModulesEnum.APACHE, os: 'linux', command: 'cd /opt/utmstack-linux-agent/beats/filebeat/ && ./filebeat modules enable apache',
      path: '/opt/utmstack-linux-agent/beats/filebeat/modules.d/', restartCmd: 'sudo systemctl restart UTMStackModulesLogsCollector'},
    {module: UtmModulesEnum.APACHE, os: 'windows', command: 'cd "C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\" '
        + '&& filebeat modules enable apache', path: 'C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\modules.d\\',
      restartCmd: 'sc stop UTMStackModulesLogsCollector && timeout /t 5 && sc start UTMStackModulesLogsCollector'},

    {module: UtmModulesEnum.ELASTICSEARCH, os: 'linux', command: 'cd /opt/utmstack-linux-agent/beats/filebeat/ && '
        + './filebeat modules enable elasticsearch',
      path: '/opt/utmstack-linux-agent/beats/filebeat/modules.d/', restartCmd: 'sudo systemctl restart UTMStackModulesLogsCollector'},
    {module: UtmModulesEnum.ELASTICSEARCH, os: 'windows', command: 'cd "C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\" '
        + '&& filebeat modules enable elasticsearch', path: 'C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\modules.d\\',
      restartCmd: 'sc stop UTMStackModulesLogsCollector && timeout /t 5 && sc start UTMStackModulesLogsCollector'},

    {module: UtmModulesEnum.HAPROXY, os: 'linux', command: 'cd /opt/utmstack-linux-agent/beats/filebeat/ && '
        + './filebeat modules enable haproxy',
      path: '/opt/utmstack-linux-agent/beats/filebeat/modules.d/', restartCmd: 'sudo systemctl restart UTMStackModulesLogsCollector'},
    {module: UtmModulesEnum.HAPROXY, os: 'windows', command: 'cd "C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\" '
        + '&& filebeat modules enable haproxy', path: 'C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\modules.d\\',
      restartCmd: 'sc stop UTMStackModulesLogsCollector && timeout /t 5 && sc start UTMStackModulesLogsCollector'},

    {module: UtmModulesEnum.MONGODB, os: 'linux', command: 'cd /opt/utmstack-linux-agent/beats/filebeat/ && '
        + './filebeat modules enable mongodb',
      path: '/opt/utmstack-linux-agent/beats/filebeat/modules.d/', restartCmd: 'sudo systemctl restart UTMStackModulesLogsCollector'},
    {module: UtmModulesEnum.MONGODB, os: 'windows', command: 'cd "C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\" '
        + '&& filebeat modules enable mongodb', path: 'C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\modules.d\\',
      restartCmd: 'sc stop UTMStackModulesLogsCollector && timeout /t 5 && sc start UTMStackModulesLogsCollector'},

    {module: UtmModulesEnum.KIBANA, os: 'linux', command: 'cd /opt/utmstack-linux-agent/beats/filebeat/ && '
        + './filebeat modules enable kibana',
      path: '/opt/utmstack-linux-agent/beats/filebeat/modules.d/', restartCmd: 'sudo systemctl restart UTMStackModulesLogsCollector'},
    {module: UtmModulesEnum.KIBANA, os: 'windows', command: 'cd "C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\" '
        + '&& filebeat modules enable kibana', path: 'C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\modules.d\\',
      restartCmd: 'sc stop UTMStackModulesLogsCollector && timeout /t 5 && sc start UTMStackModulesLogsCollector'},

    {module: UtmModulesEnum.LOGSTASH, os: 'linux', command: 'cd /opt/utmstack-linux-agent/beats/filebeat/ && '
        + './filebeat modules enable logstash',
      path: '/opt/utmstack-linux-agent/beats/filebeat/modules.d/', restartCmd: 'sudo systemctl restart UTMStackModulesLogsCollector'},
    {module: UtmModulesEnum.LOGSTASH, os: 'windows', command: 'cd "C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\" '
        + '&& filebeat modules enable logstash', path: 'C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\modules.d\\',
      restartCmd: 'sc stop UTMStackModulesLogsCollector && timeout /t 5 && sc start UTMStackModulesLogsCollector'},

    {module: UtmModulesEnum.KAFKA, os: 'linux', command: 'cd /opt/utmstack-linux-agent/beats/filebeat/ && '
        + './filebeat modules enable kafka',
      path: '/opt/utmstack-linux-agent/beats/filebeat/modules.d/', restartCmd: 'sudo systemctl restart UTMStackModulesLogsCollector'},
    {module: UtmModulesEnum.KAFKA, os: 'windows', command: 'cd "C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\" '
        + '&& filebeat modules enable kafka', path: 'C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\modules.d\\',
      restartCmd: 'sc stop UTMStackModulesLogsCollector && timeout /t 5 && sc start UTMStackModulesLogsCollector'},

    {module: UtmModulesEnum.MYSQL, os: 'linux', command: 'cd /opt/utmstack-linux-agent/beats/filebeat/ && '
        + './filebeat modules enable mysql',
      path: '/opt/utmstack-linux-agent/beats/filebeat/modules.d/', restartCmd: 'sudo systemctl restart UTMStackModulesLogsCollector'},
    {module: UtmModulesEnum.MYSQL, os: 'windows', command: 'cd "C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\" '
        + '&& filebeat modules enable mysql', path: 'C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\modules.d\\',
      restartCmd: 'sc stop UTMStackModulesLogsCollector && timeout /t 5 && sc start UTMStackModulesLogsCollector'},

    {module: UtmModulesEnum.REDIS, os: 'linux', command: 'cd /opt/utmstack-linux-agent/beats/filebeat/ && '
        + './filebeat modules enable redis',
      path: '/opt/utmstack-linux-agent/beats/filebeat/modules.d/', restartCmd: 'sudo systemctl restart UTMStackModulesLogsCollector'},
    {module: UtmModulesEnum.REDIS, os: 'windows', command: 'cd "C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\" '
        + '&& filebeat modules enable redis', path: 'C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\modules.d\\',
      restartCmd: 'sc stop UTMStackModulesLogsCollector && timeout /t 5 && sc start UTMStackModulesLogsCollector'},

    {module: UtmModulesEnum.POSTGRESQL, os: 'linux', command: 'cd /opt/utmstack-linux-agent/beats/filebeat/ && '
        + './filebeat modules enable postgresql',
      path: '/opt/utmstack-linux-agent/beats/filebeat/modules.d/', restartCmd: 'sudo systemctl restart UTMStackModulesLogsCollector'},
    {module: UtmModulesEnum.POSTGRESQL, os: 'windows', command: 'cd "C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\" '
        + '&& filebeat modules enable postgresql', path: 'C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\modules.d\\',
      restartCmd: 'sc stop UTMStackModulesLogsCollector && timeout /t 5 && sc start UTMStackModulesLogsCollector'},

    {module: UtmModulesEnum.NGINX, os: 'linux', command: 'cd /opt/utmstack-linux-agent/beats/filebeat/ && '
        + './filebeat modules enable nginx',
      path: '/opt/utmstack-linux-agent/beats/filebeat/modules.d/', restartCmd: 'sudo systemctl restart UTMStackModulesLogsCollector'},
    {module: UtmModulesEnum.NGINX, os: 'windows', command: 'cd "C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\" '
        + '&& filebeat modules enable nginx', path: 'C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\modules.d\\',
      restartCmd: 'sc stop UTMStackModulesLogsCollector && timeout /t 5 && sc start UTMStackModulesLogsCollector'},

    {module: UtmModulesEnum.AUDITD, os: 'linux', command: 'cd /opt/utmstack-linux-agent/beats/filebeat/ && '
        + './filebeat modules enable auditd',
      path: '/opt/utmstack-linux-agent/beats/filebeat/modules.d/', restartCmd: 'sudo systemctl restart UTMStackModulesLogsCollector'},

    {module: UtmModulesEnum.OSQUERY, os: 'linux', command: 'cd /opt/utmstack-linux-agent/beats/filebeat/ && '
        + './filebeat modules enable osquery',
      path: '/opt/utmstack-linux-agent/beats/filebeat/modules.d/', restartCmd: 'sudo systemctl restart UTMStackModulesLogsCollector'},
    {module: UtmModulesEnum.OSQUERY, os: 'windows', command: 'cd "C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\" '
        + '&& filebeat modules enable osquery', path: 'C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\modules.d\\',
      restartCmd: 'sc stop UTMStackModulesLogsCollector && timeout /t 5 && sc start UTMStackModulesLogsCollector'},

    {module: UtmModulesEnum.NATS, os: 'linux', command: 'cd /opt/utmstack-linux-agent/beats/filebeat/ && '
        + './filebeat modules enable nats',
      path: '/opt/utmstack-linux-agent/beats/filebeat/modules.d/', restartCmd: 'sudo systemctl restart UTMStackModulesLogsCollector'},
    {module: UtmModulesEnum.NATS, os: 'windows', command: 'cd "C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\" '
        + '&& filebeat modules enable nats', path: 'C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\modules.d\\',
      restartCmd: 'sc stop UTMStackModulesLogsCollector && timeout /t 5 && sc start UTMStackModulesLogsCollector'},
  ];
  moduleNoCommands: UtmModulesEnum[] = [];
  moduleConfigs: FilebeatModuleConfig[] = [
    {module: UtmModulesEnum.TRAEFIK, img: 'config-traefik.png'},
    {module: UtmModulesEnum.APACHE, img: 'config-apache.png'},
    {module: UtmModulesEnum.APACHE2, img: 'config-apache.png'},
    {module: UtmModulesEnum.ELASTICSEARCH, img: 'config-elasticsearch.png'},
    {module: UtmModulesEnum.HAPROXY, img: 'config-haproxy.png'},
    {module: UtmModulesEnum.IIS, img: 'config-iis-windows.png'},
    {module: UtmModulesEnum.KAFKA, img: 'config-kafka.png'},
    {module: UtmModulesEnum.LOGSTASH, img: 'config-logstash.png'},
    {module: UtmModulesEnum.MONGODB, img: 'config-mongodb.png'},
    {module: UtmModulesEnum.MYSQL, img: 'config-mysql.png'},
    {module: UtmModulesEnum.NATS, img: 'config-nats.png'},
    {module: UtmModulesEnum.NGINX, img: 'config-nginx.png'},
    {module: UtmModulesEnum.OSQUERY, img: 'config-osquery.png'},
    {module: UtmModulesEnum.POSTGRESQL, img: 'config-postgresql.png'},
    {module: UtmModulesEnum.REDIS, img: 'config-redis.png'},
    {module: UtmModulesEnum.KIBANA, img: 'config-kibana.png'},
    {module: UtmModulesEnum.AUDITD, img: 'config-auditd.png'},
  ];

  constructor() {
  }

  ngOnInit() {
  }

  getCommand(): FilebeatCommands[] {
    return this.commandsActivate.filter(value => value.module === this.filebeatModule);
  }

  getModuleConfig(): FilebeatModuleConfig {
    return this.moduleConfigs.filter(value => value.module === this.filebeatModule)[0];
  }

  showStep1(): boolean {
    switch (this.filebeatModule) {
      case (UtmModulesEnum.APACHE):
        return true;
      case (UtmModulesEnum.APACHE2):
        return true;
      case (UtmModulesEnum.AUDITD):
        return true;
      case (UtmModulesEnum.ELASTICSEARCH):
        return true;
      case (UtmModulesEnum.KIBANA):
        return true;
      case (UtmModulesEnum.LOGSTASH):
        return true;
      case (UtmModulesEnum.POSTGRESQL):
        return true;
      case (UtmModulesEnum.OSQUERY):
        return true;
      case (UtmModulesEnum.KAFKA):
        return true;
      case (UtmModulesEnum.MONGODB):
        return true;
      case (UtmModulesEnum.MYSQL):
        return true;
      case (UtmModulesEnum.TRAEFIK):
        return true;
      case (UtmModulesEnum.NATS):
        return true;
      case (UtmModulesEnum.HAPROXY):
        return true;
      case (UtmModulesEnum.NGINX):
        return true;
      case (UtmModulesEnum.REDIS):
        return true;
      default:
        return false;
    }
  }

  getFullPath(path: string, module: string) {
    return path + module + '.yml';
  }
}

export class FilebeatCommands {
  module: UtmModulesEnum;
  os: string;
  command: string;
  path: string;
  restartCmd: string;
}

export class FilebeatModuleConfig {
  module: UtmModulesEnum;
  img: string;
}
