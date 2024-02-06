import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {InfiniteScrollModule} from 'ngx-infinite-scroll';
import {UtmDashboardSharedModule} from '../../dashboard/shared/utm-dashboard-shared.module';
import {UtmSharedModule} from '../../shared/utm-shared.module';
import {IncidentResponseFilterComponent} from './component/incident-response-filter/incident-response-filter.component';
import {IncidentResponseStatusComponent} from './component/incident-response-status/incident-response-status.component';
import {IrActionCreateComponent} from './component/ir-action-create/ir-action-create.component';
import {IrCommandSelectComponent} from './component/ir-command-select/ir-command-select.component';
import {IrCreateRuleComponent} from './component/ir-create-rule/ir-create-rule.component';
import {IrExecuteCommandComponent} from './component/ir-execute-command/ir-execute-command.component';
import {IrFullLogComponent} from './component/ir-full-log/ir-full-log.component';
import {IrJobCreateComponent} from './component/ir-job-create/ir-job-create.component';
import {IraHistoryComponent} from './component/ira-history/ira-history.component';
import {IrSummaryComponent} from "./component/ir-summary/ir-summary.component";

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
    IrSummaryComponent],
  entryComponents: [IrJobCreateComponent, IrActionCreateComponent, IrCreateRuleComponent],
    exports: [IrJobCreateComponent,
        IncidentResponseFilterComponent,
        IncidentResponseStatusComponent,
        IrFullLogComponent,
        IrExecuteCommandComponent,
        IrCreateRuleComponent,
        IraHistoryComponent,
        IrSummaryComponent],
  imports: [
    CommonModule,
    NgbModule,
    UtmSharedModule,
    ReactiveFormsModule,
    InfiniteScrollModule,
    FormsModule,
    NgSelectModule,
    UtmDashboardSharedModule
  ]
})
export class IncidentResponseSharedModule {
}
