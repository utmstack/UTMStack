import {CommonModule} from '@angular/common';
import {CUSTOM_ELEMENTS_SCHEMA, NgModule, NO_ERRORS_SCHEMA} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {RouterModule} from '@angular/router';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {InfiniteScrollModule} from 'ngx-infinite-scroll';
import {InputClassResolve} from '../../shared/util/input-class-resolve';
import {UtmSharedModule} from '../../shared/utm-shared.module';
import {ComplianceSharedModule} from '../shared/compliance-shared.module';
import {CpStandardManagementComponent} from './cp-standard-management/cp-standard-management.component';
import {UtmCpExportComponent} from './utm-cp-export/utm-cp-export.component';
import {UtmCpImportComponent} from './utm-cp-import/utm-cp-import.component';
import {UtmCpReportDeleteComponent} from './utm-cp-reports/utm-cp-report-delete/utm-cp-report-delete.component';
import {UtmCpReportsComponent} from './utm-cp-reports/utm-cp-reports.component';
// tslint:disable-next-line:max-line-length
import {UtmCpStandardSectionDeleteComponent} from './utm-cp-standard-section/utm-cp-standard-section-delete/utm-cp-standard-section-delete.component';
import {UtmCpStandardSectionComponent} from './utm-cp-standard-section/utm-cp-standard-section.component';
import {UtmCpStandardDeleteComponent} from './utm-cp-standard/utm-cp-standard-delete/utm-cp-standard-delete.component';
import {UtmCpStandardComponent} from './utm-cp-standard/utm-cp-standard.component';

@NgModule({
  declarations: [
    CpStandardManagementComponent,
    UtmCpStandardComponent,
    UtmCpStandardDeleteComponent,
    UtmCpStandardSectionComponent,
    UtmCpStandardSectionDeleteComponent,
    UtmCpReportsComponent,
    UtmCpReportDeleteComponent,
    UtmCpExportComponent,
    UtmCpImportComponent],
  entryComponents: [
    UtmCpReportDeleteComponent,
    UtmCpStandardDeleteComponent,
    UtmCpStandardSectionDeleteComponent,
    UtmCpExportComponent,
    UtmCpImportComponent],
  exports: [CpStandardManagementComponent, UtmCpStandardComponent],
  providers: [InputClassResolve],
  schemas: [NO_ERRORS_SCHEMA, CUSTOM_ELEMENTS_SCHEMA],
  imports: [
    CommonModule,
    NgbModule,
    UtmSharedModule,
    ReactiveFormsModule,
    NgSelectModule,
    FormsModule,
    InfiniteScrollModule,
    ComplianceSharedModule,
    RouterModule
  ]
})
export class ComplianceManagementModule {
}
