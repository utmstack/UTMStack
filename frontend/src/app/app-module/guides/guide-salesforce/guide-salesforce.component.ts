import {Component, Input, OnInit} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {URL_REGEX} from '../../../shared/constants/regex-const';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
    selector: 'app-guide-salesforce',
    templateUrl: './guide-salesforce.component.html',
    styleUrls: ['./guide-salesforce.component.css']
})
export class GuideSalesforceComponent implements OnInit {
    @Input() integrationId: number;
    @Input() serverId: number;
    module = UtmModulesEnum;

    salesforceForm: FormGroup;

    constructor(private formBuilder: FormBuilder) {
        this.salesforceForm = this.formBuilder.group({
            clientID: ['', Validators.required],
            clientSecret: ['', Validators.required],
            username: ['', Validators.required],
            password: ['', Validators.required],
            securityToken: ['', Validators.required],
            instanceUrl: ['', [Validators.required, Validators.pattern(URL_REGEX)]],
            OAuthService: ['https://login.salesforce.com', Validators.required],
            LoginEndpoint: ['/services/oauth2/token', Validators.required],
            EventsEndPoint: ['/services/data/v57.0/sobjects/EventLogFile', Validators.required],
        });
    }

    ngOnInit() {
    }

    // mkdir -p /utmstack/sforceds && docker run --restart=always --name sforceds
    // -e "clientID=" -e "clientSecret=57653C9F1EB446253410B5AEB963537D0FE8F8E6C537A76E4C4CA9578076EE55"
    // -e "username=freddy@sandbox.com"
    // -e "securityToken=vMT8QAqfssZ3ts4OzCbmnICL"
    // -e "instanceUrl=https://quantfall-dev-ed.develop.my.salesforce.com"
    // -e "LS_JAVA_OPTS=-Xms1g -Xmx1g -Xss100m"
    // -v /utmstack/sforceds/:/local_storage
    // --log-driver json-file
    // --log-opt max-size=10m
    // --log-opt max-file=3 -d utmstack.azurecr.io/sforceds:v9


    generateDockerCommand(): string {
        const formData = this.salesforceForm.value;
        return ` mkdir -p /utmstack/sforceds && docker run --restart=always --name sforceds ` +
            `-e "clientID=${formData.clientID}" ` +
            `-e "clientSecret=<secret>${formData.clientSecret}</secret>" ` +
            `-e "username=${formData.username}" ` +
            `-e "password=<secret>${formData.password}</secret>" ` +
            `-e "securityToken=<secret>${formData.securityToken}</secret>" ` +
            `-e "instanceUrl=${formData.instanceUrl}" ` +
            `-e "LS_JAVA_OPTS=-Xms1g -Xmx1g -Xss100m" ` +
            `-v /utmstack/sforceds/:/local_storage ` +
            `--log-driver json-file --log-opt max-size=10m ` +
            `--log-opt max-file=3 -d utmstack.azurecr.io/sforceds:v9`;
    }
}
