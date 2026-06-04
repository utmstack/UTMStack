import {CommonModule} from '@angular/common';
import {CUSTOM_ELEMENTS_SCHEMA, NgModule, NO_ERRORS_SCHEMA} from '@angular/core';
import {FormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {NgxEchartsModule} from 'ngx-echarts';
import {UtmSharedModule} from '../../shared/utm-shared.module';
import {AssetCardDetailComponent} from './components/asset-card-detail/asset-card-detail.component';
import {AssetNewScanComponent} from './components/asset-new-scan/asset-new-scan.component';
import {AssetOsComponent} from './components/asset-os/asset-os.component';
// tslint:disable-next-line:max-line-length
import {AssetReportActiveFilterComponent} from './components/asset-save-report/asset-report-active-filter/asset-report-active-filter.component';
import {AssetSaveReportComponent} from './components/asset-save-report/asset-save-report.component';
import {AssetSeverityChartComponent} from './components/asset-severity-chart/asset-severity-chart.component';
import {AssetSeverityHelpComponent} from './components/asset-severity-help/asset-severity-help.component';
// tslint:disable-next-line:max-line-length
import {ScannerExportVulnerabilitiesComponent} from './components/scanner-export-vulnerabilities/scanner-export-vulnerabilities.component';
import {UsedByComponent} from './components/used-by/used-by.component';


@NgModule({
  declarations: [
    AssetSeverityChartComponent,
    AssetSeverityHelpComponent,
    UsedByComponent,
    AssetSaveReportComponent,
    AssetReportActiveFilterComponent,
    AssetCardDetailComponent,
    AssetOsComponent,
    AssetNewScanComponent,
    ScannerExportVulnerabilitiesComponent],
  entryComponents: [
    AssetSeverityChartComponent,
    AssetSeverityHelpComponent,
    UsedByComponent,
    AssetSaveReportComponent,
    AssetCardDetailComponent,
    AssetOsComponent,
    AssetNewScanComponent,
    ScannerExportVulnerabilitiesComponent],
  exports: [
    AssetSeverityChartComponent,
    AssetCardDetailComponent,
    AssetOsComponent,
    UsedByComponent,
    AssetSaveReportComponent,
    AssetCardDetailComponent
  ],
  imports: [
    CommonModule,
    NgxEchartsModule,
    FormsModule,
    NgbModule,
    UtmSharedModule,
    NgSelectModule,
  ],
  schemas: [NO_ERRORS_SCHEMA, CUSTOM_ELEMENTS_SCHEMA]
})
export class ScannerSharedModule {
}
