import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {ResizableModule} from 'angular-resizable-element';
import {IncidentResponseSharedModule} from '../incident-response/shared/incident-response-shared.module';
import {AlertIncidentStatusChangeBehavior} from '../shared/behaviors/alert-incident-status-change.behavior';
import {UtmSharedModule} from '../shared/utm-shared.module';
import {IncidentManagementComponent} from './incident-management/incident-management.component';
import {IncidentRoutingModule} from './incident-routing.module';
import {IncidentSharedModule} from './incident-shared/incident-shared.module';

@NgModule({
  declarations: [IncidentManagementComponent],
  imports: [
    CommonModule,
    IncidentRoutingModule,
    IncidentSharedModule,
    NgbModule,
    ResizableModule,
    UtmSharedModule,
    IncidentResponseSharedModule
  ],
  providers: [AlertIncidentStatusChangeBehavior]
})
export class IncidentModule {
}
