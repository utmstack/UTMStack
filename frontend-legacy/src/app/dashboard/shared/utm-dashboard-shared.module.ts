import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {GridsterModule} from 'angular-gridster2';
import {GraphicBuilderSharedModule} from '../../graphic-builder/shared/graphic-builder-shared.module';
import {UtmSharedModule} from '../../shared/utm-shared.module';
import {UtmDashboardGridComponent} from './component/utm-dashboard-grid/utm-dashboard-grid.component';
import {UtmDashboardSelectComponent} from './component/utm-dashboard-select/utm-dashboard-select.component';
import {RenderVisualizationPrintComponent} from './render-visualization-print/render-visualization-print.component';

@NgModule({
  declarations: [
    UtmDashboardGridComponent,
    UtmDashboardSelectComponent,
    RenderVisualizationPrintComponent
  ],
  exports: [
    UtmDashboardGridComponent,
    UtmDashboardSelectComponent,
    RenderVisualizationPrintComponent
  ],
  imports: [
    CommonModule,
    UtmSharedModule,
    NgbModule,
    GraphicBuilderSharedModule,
    GridsterModule
  ]
})
export class UtmDashboardSharedModule {
}
