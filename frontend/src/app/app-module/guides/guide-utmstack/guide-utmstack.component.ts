import {Component, Input, OnInit} from '@angular/core';
import {FormBuilder, FormGroup} from '@angular/forms';
import {
    FederationConnectionService
} from '../../../app-management/connection-key/shared/services/federation-connection.service';
import {GroupTypeEnum} from '../../shared/enum/group-type.enum';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';
import {Step} from '../shared/step';
import {UtmstackSteps} from './utmstack.steps';

@Component({
    selector: 'app-guide-utmstack',
    templateUrl: './guide-utmstack.component.html',
    styleUrls: ['./guide-utmstack.component.css']
})
export class GuideUtmstackComponent implements OnInit {
    @Input() integrationId: number;
    @Input() serverId: number;
    module = UtmModulesEnum;
    serverAS400FormArray: FormGroup;
    configValidity: boolean;
    groupType = GroupTypeEnum.COLLECTOR;
    steps: Step[] = UtmstackSteps;
    token: string;
    ip: string;
    vars: any;
    disablePreAction = false;
    performPreAction = true;
    architectures = [];


    constructor(private formBuilder: FormBuilder,
                private federationConnectionService: FederationConnectionService) {
    }

    ngOnInit(): void {
        this.ip = window.location.host.includes(':') ? window.location.host.split(':')[0] : window.location.host;
        this.getToken();
    }

    getToken() {
        this.federationConnectionService.getToken().subscribe(response => {
            if (response.body !== null && response.body !== '') {
                this.token = response.body;
            } else {
                this.token = '';
            }
            this.vars = {
                V_IP: this.ip,
                V_TOKEN: this.token
            };
        });
    }

  configValidChange($event: boolean) {
    this.configValidity = !$event;
  }

  onDisable() {
    this.disablePreAction = true;
  }
}
