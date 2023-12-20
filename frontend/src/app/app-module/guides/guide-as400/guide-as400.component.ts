import {Component, Input, OnInit} from '@angular/core';
import {FormArray, FormBuilder, FormGroup, Validators} from '@angular/forms';
import * as moment from 'moment/moment';
import {IP_HOSTNAME_REGEX} from '../../../shared/constants/regex-const';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
    selector: 'app-guide-as400',
    templateUrl: './guide-as400.component.html',
    styleUrls: ['./guide-as400.component.css']
})
export class GuideAs400Component implements OnInit {
    @Input() integrationId: number;
    @Input() serverId: number;
    module = UtmModulesEnum;
    serverAS400FormArray: FormGroup;

    constructor(private formBuilder: FormBuilder) {
        this.serverAS400FormArray = this.formBuilder.group({
            serversAS400: this.formBuilder.array([
                this.createAS400ServerFormGroup(),
            ]),
        });
    }

    ngOnInit(): void {
    }

    get serverAS400Controls() {
        return this.serverAS400FormArray.controls.serversAS400 as FormArray;
    }

    addServer() {
        this.serverAS400Controls.push(this.createAS400ServerFormGroup());
    }

    removeServer(index: number) {
        const servers = this.serverAS400FormArray.get('serversAS400') as FormArray;
        servers.removeAt(index);
    }

    private createAS400ServerFormGroup() {
        return this.formBuilder.group({
            hostname: ['', [Validators.required, Validators.pattern(IP_HOSTNAME_REGEX)]],
            userId: ['', Validators.required],
            user_password: ['', Validators.required],
            tenant: ['', Validators.required],
        });
    }

    downloadConfig() {
        const dataStr = JSON.stringify(this.serverAS400FormArray.value);
        const dataUri = 'data:application/json;charset=utf-8,' + encodeURIComponent(dataStr);
        const exportFileDefaultName = 'as400-' + moment(new Date()).format('YYYY-MM-DD') + '.json';
        const linkElement = document.createElement('a');
        linkElement.setAttribute('href', dataUri);
        linkElement.setAttribute('download', exportFileDefaultName);
        linkElement.click();
    }


    // docker run --restart=always --name as400jds
    // -e "SYSLOG_PROTOCOL=udp"
    // -e "SYSLOG_HOST=10.21.199.1"
    // -e "SYSLOG_PORT=514"
    // -e "LS_JAVA_OPTS=-Xms1g -Xmx1g -Xss100m"
    // -v /utmstack/as400jds/:/local_storage
    // --log-driver json-file
    // --log-opt max-size=10m
    // --log-opt max-file=3
    // -d utmstack.azurecr.io/as400jds:v9
    generateDockerCommand(): string {
        return `docker run --restart=always --name as400jds ` +
            `-e "SYSLOG_PROTOCOL=udp" ` +
            `-e "SYSLOG_HOST=10.21.199.1" ` +
            `-e "SYSLOG_PORT=514" ` +
            `-e "LS_JAVA_OPTS=-Xms1g -Xmx1g -Xss100m" ` +
            `-v /utmstack/as400jds/:/local_storage ` +
            `--log-driver json-file --log-opt max-size=10m ` +
            `--log-opt max-file=3 ` +
            `-d utmstack.azurecr.io/as400jds:v9`;
    }
}
