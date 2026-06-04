import {Component, Input, OnInit} from '@angular/core';
import {FormBuilder, FormGroup} from '@angular/forms';
import {
    FederationConnectionService
} from '../../../app-management/connection-key/shared/services/federation-connection.service';
import {GroupTypeEnum} from '../../shared/enum/group-type.enum';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';
import {Step} from '../shared/step';
import {ACTIONS, PLATFORM} from './constants';
import {OracleSteps} from './oracle.steps';

@Component({
    selector: 'app-guide-oracle',
    templateUrl: './oracle.component.html',
    styleUrls: ['./oracle.component.css']
})
export class OracleComponent implements OnInit {
    @Input() integrationId: number;
    @Input() serverId: number;
    module = UtmModulesEnum;
    serverAS400FormArray: FormGroup;
    configValidity: boolean;
    groupType = GroupTypeEnum.COLLECTOR;
    platforms = PLATFORM;
    actions = ACTIONS;
    steps: Step[] = OracleSteps;
    token: string;
    ip: string;
    vars: any;
    disablePreAction = false;
    performPreAction = true;
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
    console.log('Disabled');
    this.disablePreAction = true;
  }

  onRunDisable($event: any) {
    this.performPreAction = $event;
    console.log(this.performPreAction);
  }
}
