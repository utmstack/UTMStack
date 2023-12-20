import {CommonModule} from '@angular/common';
import {CUSTOM_ELEMENTS_SCHEMA, NgModule, NO_ERRORS_SCHEMA} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {TranslateModule} from '@ngx-translate/core';
import {ResizableModule} from 'angular-resizable-element';
import {NgxJsonViewerModule} from 'ngx-json-viewer';
import {IncidentSharedModule} from '../../incident/incident-shared/incident-shared.module';
import {AlertIncidentStatusChangeBehavior} from '../../shared/behaviors/alert-incident-status-change.behavior';
import {NewAlertBehavior} from '../../shared/behaviors/new-alert.behavior';
import {UtmSharedModule} from '../../shared/utm-shared.module';
import {AlertFullDetailComponent} from './alert-full-detail/alert-full-detail.component';
import {AlertManagementRouting} from './alert-management-routing.module';
import {AlertReportViewComponent} from './alert-report-view/alert-report-view.component';
import {AlertReportsComponent} from './alert-reports/alert-reports.component';
import {AlertReportFilterComponent} from './alert-reports/shared/components/alert-report-filter/alert-report-filter.component';
import {SaveAlertReportComponent} from './alert-reports/shared/components/save-report/save-report.component';
import {AlertViewComponent} from './alert-view/alert-view.component';
import {AlertManagementSharedModule} from './shared/alert-management-shared.module';

@NgModule({
  declarations: [
    AlertViewComponent,
    AlertReportsComponent,
    SaveAlertReportComponent,

    AlertReportFilterComponent,
    AlertReportViewComponent,
    AlertFullDetailComponent,
  ],
  imports: [
    CommonModule,
    AlertManagementRouting,
    UtmSharedModule,
    FormsModule,
    NgbModule,
    AlertManagementSharedModule,
    NgxJsonViewerModule,
    ResizableModule,
    TranslateModule,
    NgSelectModule,
    ReactiveFormsModule,
    IncidentSharedModule
  ],
  entryComponents: [
    SaveAlertReportComponent,
    AlertReportFilterComponent],
  providers: [NewAlertBehavior, AlertIncidentStatusChangeBehavior],
  exports: [],
  schemas: [NO_ERRORS_SCHEMA, CUSTOM_ELEMENTS_SCHEMA]
})
export class AlertManagementModule {
}
