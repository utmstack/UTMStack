import {CommonModule} from '@angular/common';
import {CUSTOM_ELEMENTS_SCHEMA, NgModule, NO_ERRORS_SCHEMA} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {UtmDashboardSharedModule} from '../../dashboard/shared/utm-dashboard-shared.module';
import {UtmSharedModule} from '../../shared/utm-shared.module';
import {UtmComplianceCreateComponent} from './components/utm-compliance-create/utm-compliance-create.component';
import {UtmCpStSectionSelectComponent} from './components/utm-cp-st-section-select/utm-cp-st-section-select.component';
import {UtmCpStandardCreateComponent} from './components/utm-cp-standard-create/utm-cp-standard-create.component';
import {UtmCpStandardSectionCreateComponent} from './components/utm-cp-standard-section-create/utm-cp-standard-section-create.component';
import {UtmCpStandardSelectComponent} from './components/utm-cp-standard-select/utm-cp-standard-select.component';
import {UtmReportInfoViewComponent} from './components/utm-report-info-view/utm-report-info-view.component';
import {UtmSaveAsComplianceComponent} from './components/utm-save-as-compliance/utm-save-as-compliance.component';

@NgModule({
  declarations: [
    UtmSaveAsComplianceComponent,
    UtmCpStandardSelectComponent,
    UtmCpStSectionSelectComponent,
    UtmCpStandardCreateComponent,
    UtmCpStandardSectionCreateComponent,
    UtmReportInfoViewComponent,
    UtmComplianceCreateComponent
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
    UtmComplianceCreateComponent
  ],
  exports: [
    UtmSaveAsComplianceComponent,
    UtmCpStandardSelectComponent,
    UtmCpStSectionSelectComponent,
    UtmCpStandardCreateComponent,
    UtmCpStandardSectionCreateComponent,
    UtmReportInfoViewComponent
  ]
})
export class ComplianceSharedModule {
}
