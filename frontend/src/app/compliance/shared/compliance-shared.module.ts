import {CommonModule} from '@angular/common';
import {CUSTOM_ELEMENTS_SCHEMA, NgModule, NO_ERRORS_SCHEMA} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {UtmDashboardSharedModule} from '../../dashboard/shared/utm-dashboard-shared.module';
import {UtmSharedModule} from '../../shared/utm-shared.module';
import {UtmComplianceCreateComponent} from './components/utm-compliance-create/utm-compliance-create.component';
import {UtmCpCronEditorComponent} from './components/utm-cp-cron-editor/utm-cp-cron-editor.component';
import {UtmCpStSectionSelectComponent} from './components/utm-cp-st-section-select/utm-cp-st-section-select.component';
import {UtmCpStandardCreateComponent} from './components/utm-cp-standard-create/utm-cp-standard-create.component';
import {UtmCpStandardSectionCreateComponent} from './components/utm-cp-standard-section-create/utm-cp-standard-section-create.component';
import {UtmCpStandardSelectComponent} from './components/utm-cp-standard-select/utm-cp-standard-select.component';
import {UtmReportInfoViewComponent} from './components/utm-report-info-view/utm-report-info-view.component';
import {UtmSaveAsComplianceComponent} from './components/utm-save-as-compliance/utm-save-as-compliance.component';
import {
  UtmComplianceScheduleCreateComponent
} from "./components/utm-compliance-schedule-create/utm-compliance-schedule-create.component";
import {UtmComplianceSelectComponent} from "./components/utm-compliance-select/utm-compliance-select.component";
import {
  UtmComplianceScheduleDeleteComponent
} from "./components/utm-compliance-schedule-delete/utm-compliance-schedule-delete.component";


@NgModule({
  declarations: [
    UtmSaveAsComplianceComponent,
    UtmCpStandardSelectComponent,
    UtmCpStSectionSelectComponent,
    UtmCpStandardCreateComponent,
    UtmCpStandardSectionCreateComponent,
    UtmReportInfoViewComponent,
    UtmComplianceCreateComponent,
    UtmComplianceScheduleCreateComponent,
    UtmCpCronEditorComponent,
    UtmComplianceSelectComponent,
    UtmComplianceScheduleDeleteComponent
  ],
  imports: [
    CommonModule,
    UtmSharedModule,
    NgSelectModule,
    FormsModule,
    ReactiveFormsModule,
    NgbModule,
    UtmDashboardSharedModule
  ],
  schemas: [NO_ERRORS_SCHEMA, CUSTOM_ELEMENTS_SCHEMA],
  entryComponents: [
    UtmSaveAsComplianceComponent,
    UtmCpStandardSelectComponent,
    UtmCpStSectionSelectComponent,
    UtmCpStandardCreateComponent,
    UtmCpStandardSectionCreateComponent,
    UtmComplianceCreateComponent,
    UtmComplianceScheduleCreateComponent,
    UtmComplianceScheduleDeleteComponent
  ],
  exports: [
    UtmSaveAsComplianceComponent,
    UtmCpStandardSelectComponent,
    UtmCpStSectionSelectComponent,
    UtmCpStandardCreateComponent,
    UtmCpStandardSectionCreateComponent,
    UtmReportInfoViewComponent,
    UtmComplianceScheduleCreateComponent,
    UtmComplianceScheduleDeleteComponent
  ]
})
export class ComplianceSharedModule {
}
