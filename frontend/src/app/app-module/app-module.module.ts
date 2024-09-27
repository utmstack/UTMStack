import {CommonModule} from '@angular/common';
import {CUSTOM_ELEMENTS_SCHEMA, NgModule, NO_ERRORS_SCHEMA} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {InlineSVGModule} from 'ng-inline-svg';
import {UtmSharedModule} from '../shared/utm-shared.module';
import {AppModuleRoutingModuleRouting} from './app-module-routing.module';
import {AppModuleViewComponent} from './app-module-view/app-module-view.component';
import {IntAwsGroupsComponent} from './conf/int-aws-groups/int-aws-groups.component';
import {IntAzureEventhubComponent} from './conf/int-azure-eventhub/int-azure-eventhub.component';
import {IntGenericConfigInputComponent} from './conf/int-generic-config-input/int-generic-config-input.component';
import {IntGenericGroupConfigComponent} from './conf/int-generic-group-config/int-generic-group-config.component';
import {IntGoogleProjectsComponent} from './conf/int-google-projects/int-google-projects.component';
import {IntLdapGroupsComponent} from './conf/int-ldap-groups/int-ldap-groups.component';
import { GuideAs400Component } from './guides/guide-as400/guide-as400.component';
import {GuideAssetScannerComponent} from './guides/guide-asset-scanner/guide-asset-scanner.component';
import {GuideAwsBeanstalkComponent} from './guides/guide-aws-beanstalk/guide-aws-beanstalk.component';
import {GuideAwsCloudtrailComponent} from './guides/guide-aws-cloudtrail/guide-aws-cloudtrail.component';
import {GuideAwsEcsFargateComponent} from './guides/guide-aws-ecs-fargate/guide-aws-ecs-fargate.component';
import {GuideAwsIamUserComponent} from './guides/guide-aws-iam-user/guide-aws-iam-user.component';
import {GuideAwsLambdaComponent} from './guides/guide-aws-lambda/guide-aws-lambda.component';
import {GuideAwsRdsMsSqlComponent} from './guides/guide-aws-rds-ms-sql/guide-aws-rds-ms-sql.component';
import {GuideAwsRdsPostgresComponent} from './guides/guide-aws-rds-postgres/guide-aws-rds-postgres.component';
import {GuideAwsTrafficMirrorComponent} from './guides/guide-aws-traffic-mirror/guide-aws-traffic-mirror.component';
import {GuideAzureComponent} from './guides/guide-azure/guide-azure.component';
import { GuideBitdefenderComponent } from './guides/guide-bitdefender/guide-bitdefender.component';
import { GuideCiscoComponent } from './guides/guide-cisco/guide-cisco.component';
import { GuideEsetComponent } from './guides/guide-eset/guide-eset.component';
import {GuideFileClasificationComponent} from './guides/guide-file-clasification/guide-file-clasification.component';
import { GuideFilebeatGenericComponent } from './guides/guide-filebeat-generic/guide-filebeat-generic.component';
import {GuideFilebeatComponent} from './guides/guide-filebeat/guide-filebeat.component';
import {GuideGitHubComponent} from './guides/guide-github/guide-github.component';
import {GuideGooglePubsubComponent} from './guides/guide-google-pubsub/guide-google-pubsub.component';
import {GuideIisComponent} from './guides/guide-iis/guide-iis.component';
import { GuideJsonComponent } from './guides/guide-json/guide-json.component';
import { GuideKasperskyComponent } from './guides/guide-kaspersky/guide-kaspersky.component';
import {GuideLdapComponent} from './guides/guide-ldap/guide-ldap.component';
import { GuideLinuxAgentComponent } from './guides/guide-linux-agent/guide-linux-agent.component';
import { GuideMacosAgentComponent } from './guides/guide-macos-agent/guide-macos-agent.component';
import {GuideNetflowComponent} from './guides/guide-netflow/guide-netflow.component';
import {GuideOffice365Component} from './guides/guide-office365/guide-office365.component';
import { GuideSalesforceComponent } from './guides/guide-salesforce/guide-salesforce.component';
import { GuideSentinelOneComponent } from './guides/guide-sentinel-one/guide-sentinel-one.component';
import { GuideSocAiComponent } from './guides/guide-soc-ai/guide-soc-ai.component';
import {GuideSophosComponent} from './guides/guide-sophos/guide-sophos.component';
import {GuideSyslogComponent} from './guides/guide-syslog/guide-syslog.component';
import {GuideVmwareSyslogComponent} from './guides/guide-vmware-syslog/guide-vmware-syslog.component';
import {GuideVulnerabilitiesComponent} from './guides/guide-vulnerabilities/guide-vulnerabilities.component';
import {GuideWebrootComponent} from './guides/guide-webroot/guide-webroot.component';
import {GuideWindowFaaComponent} from './guides/guide-window-faa/guide-window-faa.component';
import {GuideWinlogbeatComponent} from './guides/guide-winlogbeat/guide-winlogbeat.component';
import {StepComponent, StepDirective} from './guides/shared/components/step.component';
import {UtmListComponent} from './guides/shared/components/utm-list.component';
import {ModuleIntegrationComponent} from './module-integration/module-integration.component';
import {AppModuleSharedModule} from './shared/app-module-shared.module';
import {AgentActionCommandComponent} from './guides/shared/components/agent-action-command.component';
import {InstallLogCollectorComponent} from './guides/shared/components/install-log-collector.component';
import {ModuleResolverService} from "./services/module.resolver.service";
import {GenericConfiguration} from "./conf/int-generic-group-config/int-config-types/generic-configuration";
import {CollectorConfiguration} from "./conf/int-generic-group-config/int-config-types/collector-configuration";
import {IntegrationConfigFactory} from "./conf/int-generic-group-config/int-config-types/IntegrationConfigFactory";


@NgModule({
  declarations: [AppModuleViewComponent,
    IntAwsGroupsComponent,
    IntGenericConfigInputComponent,
    IntLdapGroupsComponent,
    GuideWinlogbeatComponent,
    GuideAwsIamUserComponent,
    GuideAwsBeanstalkComponent,
    GuideAwsEcsFargateComponent,
    GuideAwsLambdaComponent,
    GuideAwsRdsMsSqlComponent,
    GuideAwsRdsPostgresComponent,
    GuideAzureComponent,
    GuideAwsCloudtrailComponent,
    GuideOffice365Component,
    GuideIisComponent,
    GuideFilebeatComponent,
    GuideVmwareSyslogComponent,
    GuideAwsTrafficMirrorComponent,
    IntGenericConfigInputComponent,
    IntAwsGroupsComponent,
    IntLdapGroupsComponent,
    GuideLdapComponent,
    GuideWebrootComponent,
    GuideSyslogComponent,
    GuideWindowFaaComponent,
    GuideNetflowComponent,
    GuideSophosComponent,
    ModuleIntegrationComponent,
    IntGoogleProjectsComponent,
    IntAzureEventhubComponent,
    GuideGooglePubsubComponent,
    GuideAssetScannerComponent,
    GuideVulnerabilitiesComponent,
    GuideFileClasificationComponent,
    IntGenericGroupConfigComponent,
    GuideMacosAgentComponent,
    GuideLinuxAgentComponent,
    GuideJsonComponent,
    GuideFilebeatGenericComponent,
    GuideCiscoComponent,
    GuideEsetComponent,
    GuideKasperskyComponent,
    GuideSentinelOneComponent,
    GuideGitHubComponent,
    GuideSalesforceComponent,
    GuideBitdefenderComponent,
    GuideAs400Component,
    GuideSocAiComponent,
    UtmListComponent,
    StepComponent,
    StepDirective,
    AgentActionCommandComponent,
    InstallLogCollectorComponent
  ],
  imports: [
    CommonModule,
    AppModuleRoutingModuleRouting,
    AppModuleSharedModule,
    UtmSharedModule,
    InlineSVGModule,
    FormsModule,
    NgbModule,
    ReactiveFormsModule,
    NgSelectModule,
  ],
  providers: [
    ModuleResolverService,
    GenericConfiguration,
    CollectorConfiguration,
    IntegrationConfigFactory
  ],
  entryComponents: [],
  exports: [],
  schemas: [CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA]
})
export class AppModuleModule {
}
