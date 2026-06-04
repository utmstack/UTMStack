import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';
import {FilebeatCommands} from "../guide-filebeat-generic/guide-filebeat-generic.component";

@Component({
  selector: 'app-guide-filebeat',
  templateUrl: './guide-filebeat.component.html',
  styleUrls: ['./guide-filebeat.component.scss']
})
export class GuideFilebeatComponent implements OnInit {
  @Input() integrationId: number;
  @Input() serverId: number;
  module = UtmModulesEnum;
  commandsRsl: RsyslogCommands[] = [
    {manager: 'Using APT', command: 'sudo apt-get update <br> sudo apt-get install rsyslog'},
    {manager: 'Using YUM', command: 'sudo yum install rsyslog'},
    {manager: 'Using DNF', command: 'sudo dnf install rsyslog'},
  ];

  constructor() {
  }

  ngOnInit() {
  }

  getCommand(): RsyslogCommands[] {
    return this.commandsRsl;
  }

}

export class RsyslogCommands {
  manager: string;
  command: string;
}
