import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {TranslateModule} from '@ngx-translate/core';
import {AlertManagementSharedModule} from '../data-management/alert-management/shared/alert-management-shared.module';
import {UtmSharedModule} from '../shared/utm-shared.module';
import {
  IncidentResponseAutomationComponent
} from './incident-response-automation/incident-response-automation.component';
import {
  IrVariableCreateComponent
} from './incident-response-variables/ir-variable-create/ir-variable-create.component';
import {IncidentResponseCommandComponent} from './incident-response-command/incident-response-command.component';
import {IncidentResponseRoutingModule} from './incident-response-routing.module';
import {IncidentResponseVariablesComponent} from './incident-response-variables/incident-response-variables.component';
import {IncidentResponseViewComponent} from './incident-response-view/incident-response-view.component';
import {IncidentResponseSharedModule} from './shared/incident-response-shared.module';

@NgModule({
  declarations: [IncidentResponseViewComponent, IncidentResponseCommandComponent,
    IncidentResponseAutomationComponent, IncidentResponseVariablesComponent, IrVariableCreateComponent],
  imports: [
    CommonModule,
    IncidentResponseRoutingModule,
    IncidentResponseSharedModule,
    UtmSharedModule,
    NgbModule,
    NgSelectModule,
    FormsModule,
    AlertManagementSharedModule,
    TranslateModule,
    ReactiveFormsModule
  ], entryComponents: [IrVariableCreateComponent]
})
export class IncidentResponseModule {
}
