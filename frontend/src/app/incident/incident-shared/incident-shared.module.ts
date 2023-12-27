import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {AlertManagementSharedModule} from '../../data-management/alert-management/shared/alert-management-shared.module';
import {AlertIncidentStatusChangeBehavior} from '../../shared/behaviors/alert-incident-status-change.behavior';
import {UtmSharedModule} from '../../shared/utm-shared.module';
import {AddAlertToIncidentComponent} from './component/add-alert-to-incident/add-alert-to-incident.component';
import {CreateIncidentComponent} from './component/create-incident/create-incident.component';
import {IncidentHistoryComponent} from './component/incident-history/incident-history.component';
import {IncidentNotesComponent} from './component/incident-notes/incident-notes.component';
import {IncidentRelatedAlertComponent} from './component/incident-related-alert/incident-related-alert.component';


@NgModule({
  declarations: [CreateIncidentComponent, IncidentHistoryComponent,
    IncidentNotesComponent, AddAlertToIncidentComponent,
    IncidentRelatedAlertComponent,
    ],
  imports: [
    CommonModule,
    NgbModule,
    UtmSharedModule,
    ReactiveFormsModule,
    NgSelectModule,
    FormsModule,
    AlertManagementSharedModule,
  ],
  exports: [CreateIncidentComponent, IncidentHistoryComponent,
    IncidentNotesComponent, AddAlertToIncidentComponent,
    IncidentRelatedAlertComponent,
    ],
  entryComponents: [CreateIncidentComponent, AddAlertToIncidentComponent],
  providers: [AlertIncidentStatusChangeBehavior]
})
export class IncidentSharedModule {
}
