import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {TranslateModule} from '@ngx-translate/core';
import {ResizableModule} from 'angular-resizable-element';
import {NgxEchartsModule} from 'ngx-echarts';
import {InfiniteScrollModule} from 'ngx-infinite-scroll';
import {NgxJsonViewerModule} from 'ngx-json-viewer';
import {IncidentSharedModule} from '../../incident/incident-shared/incident-shared.module';
import {UtmSharedModule} from '../../shared/utm-shared.module';
import {AlertManagementSharedModule} from '../alert-management/shared/alert-management-shared.module';
import { AdversaryAlertsGraphComponent } from './adversary-alerts-graph/adversary-alerts-graph.component';
import {AdversaryManagementRoutingModule} from './adversary-management-routing.module';
import {AdversaryViewComponent} from './adversary-view/adversary-view.component';
import {AdversaryAlertService} from './services/adversary-alert-.service';

@NgModule({
  declarations: [
    AdversaryViewComponent,
    AdversaryAlertsGraphComponent
  ],
  imports: [
    CommonModule,
    UtmSharedModule,
    FormsModule,
    NgbModule,
    AlertManagementSharedModule,
    NgxJsonViewerModule,
    ResizableModule,
    TranslateModule,
    NgSelectModule,
    ReactiveFormsModule,
    IncidentSharedModule,
    InfiniteScrollModule,
    AdversaryManagementRoutingModule,
    NgxEchartsModule
  ],
  providers: [
    AdversaryAlertService
  ]
})
export class AdversaryManagementModule { }
