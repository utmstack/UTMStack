import {Component, Input, OnInit} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {IP_HOSTNAME_REGEX, URL_REGEX} from '../../../shared/constants/regex-const';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-guide-bitdefender',
  templateUrl: './guide-bitdefender.component.html',
  styleUrls: ['./guide-bitdefender.component.css']
})
export class GuideBitdefenderComponent implements OnInit {
  @Input() integrationId: number;
  @Input() serverId: number;
  module = UtmModulesEnum;

  bfForm: FormGroup;
  configValidity = false;

  constructor(private formBuilder: FormBuilder) {
  }

  ngOnInit(): void {
    this.bfForm = this.formBuilder.group({
      syslogProtocol: ['tcp', Validators.required],
      syslogHost: ['10.21.199.1', Validators.required],
      syslogPort: ['514', Validators.required],
      connectorPort: ['8000', Validators.required],
      bdgzApiKey: ['', Validators.required],
      accessUrl: ['', [Validators.required, Validators.pattern(URL_REGEX)]],
      panelUrl: ['', [Validators.required,
        Validators.pattern(IP_HOSTNAME_REGEX)]
      ],
    });
  }

  generateDockerCommand(): string {
    return `docker run --restart=always --name bitdefender ` +
      `-e "SYSLOG_PROTOCOL=${this.bfForm.value.syslogProtocol}" ` +
      `-e "SYSLOG_HOST=${this.bfForm.value.syslogHost}" ` +
      `-e "SYSLOG_PORT=${this.bfForm.value.syslogPort}" ` +
      `-e "CONNECTOR_PORT=${this.bfForm.value.connectorPort}" ` +
      `-e "BDGZ_API_KEY=<secret>${this.bfForm.value.bdgzApiKey}</secret>" ` +
      `-e "BDGZ_ACCESS_URL=${this.bfForm.value.accessUrl}" ` +
      `-e "BDGZ_URL=https://${this.bfForm.value.panelUrl}" ` +
      `--log-driver json-file --log-opt max-size=10m ` +
      `--log-opt max-file=3 -p 8000:8000 ` +
      `-d utmstack.azurecr.io/bitdefender:v9`;
  }

  configValidChange($event: boolean) {
    this.configValidity = !$event;
  }

}
