import {CommonModule} from '@angular/common';
import {CUSTOM_ELEMENTS_SCHEMA, NgModule} from '@angular/core';
import {NgSelectModule} from '@ng-select/ng-select';
import {GridsterModule} from 'angular-gridster2';
import {AppModuleSharedModule} from '../app-module/shared/app-module-shared.module';
import {UtmstackCoreModule} from '../core/core.module';
import {UtmDefinedChartsModule} from '../defined-charts/utm-defined-charts.module';
import {GraphicBuilderSharedModule} from '../graphic-builder/shared/graphic-builder-shared.module';
import {UtmSharedModule} from '../shared/utm-shared.module';
import {VulnerabilitySharedModule} from '../vulnerability-scanner/vulnerability-shared/vulnerability-shared.module';
import {ComplianceExportComponent} from './compliance-export/compliance-export.component';
import {DashboardExportCustomComponent} from './dashboard-export-custom/dashboard-export-custom.component';
import {DashboardExportPdfComponent} from './dashboard-export-pdf/dashboard-export-pdf.component';
import {DashboardExportPreviewComponent} from './dashboard-export-preview/dashboard-export-preview.component';
import {DashboardLogSourcesComponent} from './dashboard-log-sources/dashboard-log-sources.component';
import {DashboardOverviewComponent} from './dashboard-overview/dashboard-overview.component';
import {DashboardRenderComponent} from './dashboard-render/dashboard-render.component';
import {DashboardRoutingModule} from './dashboard-routing.module';
import {DashboardViewComponent} from './dashboard-view/dashboard-view.component';
import {ReportExportComponent} from './report-export/report-export.component';
import {UtmDashboardSharedModule} from './shared/utm-dashboard-shared.module';

@NgModule({
  declarations: [DashboardViewComponent,
    DashboardRenderComponent,
    DashboardOverviewComponent,
    DashboardExportPreviewComponent,
    DashboardExportCustomComponent,
    DashboardExportPdfComponent,
    DashboardLogSourcesComponent,
    ComplianceExportComponent,
    ReportExportComponent],
  imports: [
    CommonModule,
    AppModuleSharedModule,
    DashboardRoutingModule,
    UtmstackCoreModule,
    UtmSharedModule,
    NgSelectModule,
    UtmDefinedChartsModule,
    UtmDashboardSharedModule,
    GraphicBuilderSharedModule,
    VulnerabilitySharedModule,
    GridsterModule
  ],
  entryComponents: [DashboardExportPreviewComponent],
  exports: [
    DashboardViewComponent
  ],
  schemas: [CUSTOM_ELEMENTS_SCHEMA]
})
export class UtmDashboardModule {

}
