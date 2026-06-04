import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {InfiniteScrollModule} from 'ngx-infinite-scroll';
import {UtmDashboardSharedModule} from '../../dashboard/shared/utm-dashboard-shared.module';
import {UtmSharedModule} from '../../shared/utm-shared.module';
import { ActionBuilderComponent } from './component/action-builder/action-builder.component';
import { ActionConditionalComponent } from './component/action-conditional/action-conditional.component';
import { ActionSidebarComponent } from './component/action-sidebar/action-sidebar.component';
import { ActionTerminalComponent } from './component/action-terminal/action-terminal.component';
import { ConditionBuilderComponent } from './component/condition-builder/condition-builder.component';
import { ConditionItemComponent } from './component/condition-item/condition-item.component';
import {IncidentResponseFilterComponent} from './component/incident-response-filter/incident-response-filter.component';
import {IncidentResponseStatusComponent} from './component/incident-response-status/incident-response-status.component';
import {IrActionCreateComponent} from './component/ir-action-create/ir-action-create.component';
import {IrCommandSelectComponent} from './component/ir-command-select/ir-command-select.component';
import {IrCreateRuleComponent} from './component/ir-create-rule/ir-create-rule.component';
import {IrExecuteCommandComponent} from './component/ir-execute-command/ir-execute-command.component';
import {IrFullLogComponent} from './component/ir-full-log/ir-full-log.component';
import {IrJobCreateComponent} from './component/ir-job-create/ir-job-create.component';
import {IrSummaryComponent} from './component/ir-summary/ir-summary.component';
import {IraHistoryComponent} from './component/ira-history/ira-history.component';
import {NewPlaybookComponent} from './component/new-playbook/new-playbook.component';
import { AgentSidebarComponent } from './component/agent-sidebar/agent-sidebar.component';
import { AgentInfoComponent } from './component/agent-info/agent-info.component';
import { InteractiveConsoleComponent } from './component/interactive-console/interactive-console.component';
import {PlaybookService} from "./services/playbook.service";
import {RouterModule} from "@angular/router";

@NgModule({
  declarations: [IrJobCreateComponent,
    IrActionCreateComponent,
    IncidentResponseFilterComponent,
    IncidentResponseStatusComponent,
    IrFullLogComponent,
    IrExecuteCommandComponent,
    IrCommandSelectComponent,
    IrCreateRuleComponent,
    IraHistoryComponent,
    IrSummaryComponent,
    ConditionBuilderComponent,
    ConditionItemComponent,
    ActionBuilderComponent,
    ActionSidebarComponent,
    ActionTerminalComponent,
    ActionConditionalComponent,
    NewPlaybookComponent,
    AgentSidebarComponent,
    AgentInfoComponent,
    InteractiveConsoleComponent],

  entryComponents: [
    IrJobCreateComponent,
    IrActionCreateComponent,
    IrCreateRuleComponent,
    ActionTerminalComponent,
    NewPlaybookComponent
  ],

  exports: [
    IrJobCreateComponent,
    IncidentResponseFilterComponent,
    IncidentResponseStatusComponent,
    IrFullLogComponent,
    IrExecuteCommandComponent,
    IrCreateRuleComponent,
    IraHistoryComponent,
    IrSummaryComponent,
    ConditionBuilderComponent,
    ActionBuilderComponent,
    ActionSidebarComponent,
    ActionTerminalComponent,
    ActionConditionalComponent,
    NewPlaybookComponent,
    AgentSidebarComponent,
    AgentInfoComponent,
    InteractiveConsoleComponent
  ],
  imports: [
    CommonModule,
    NgbModule,
    UtmSharedModule,
    ReactiveFormsModule,
    InfiniteScrollModule,
    FormsModule,
    NgSelectModule,
    UtmDashboardSharedModule,
    RouterModule
  ]
})
export class IncidentResponseSharedModule {
}
