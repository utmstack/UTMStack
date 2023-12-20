import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {RouterModule} from '@angular/router';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {UtmDashboardSharedModule} from '../dashboard/shared/utm-dashboard-shared.module';
import {UtmSharedModule} from '../shared/utm-shared.module';
import {ReportCustomViewComponent} from './report-custom-view/report-custom-view.component';
import {ReportRoutingModule} from './report-routing.module';
import {ReportTemplateResultComponent} from './report-template-result/report-template-result.component';
import {ReportTemplateComponent} from './report-template/report-template.component';
import {ReportCustomExportModalComponent} from './shared/component/report-custom-export-modal/report-custom-export-modal.component';
import {ReportParamDateLimitComponent} from './shared/component/report-param-date-limit/report-param-date-limit.component';
import {ReportParamDateRangeComponent} from './shared/component/report-param-date-range/report-param-date-range.component';
import {ReportParamFileIntegrityComponent} from './shared/component/report-param-file-integrity/report-param-file-integrity.component';
import {ReportParamRiskAssetmentComponent} from './shared/component/report-param-risk-assetment/report-param-risk-assetment.component';
import {UtmReportCreateComponent} from './shared/component/utm-report-create/utm-report-create.component';
import {UtmReportSectionCreateComponent} from './shared/component/utm-report-section-create/utm-report-section-create.component';

@NgModule({
  declarations: [ReportTemplateComponent,
    UtmReportCreateComponent,
    UtmReportSectionCreateComponent,
    ReportTemplateResultComponent,
    ReportParamRiskAssetmentComponent,
    ReportCustomExportModalComponent,
    ReportParamDateRangeComponent,
    ReportParamDateLimitComponent,
    ReportParamFileIntegrityComponent,
    ReportCustomViewComponent],
  entryComponents: [UtmReportCreateComponent,
    UtmReportSectionCreateComponent,
    ReportCustomExportModalComponent],
  imports: [
    CommonModule,
    ReportRoutingModule,
    UtmSharedModule,
    RouterModule,
    UtmDashboardSharedModule,
    FormsModule,
    NgSelectModule,
    ReactiveFormsModule,
    NgbModule
  ]
})
export class ReportModule {
}
