import {CommonModule} from '@angular/common';
import {CUSTOM_ELEMENTS_SCHEMA, NgModule, NO_ERRORS_SCHEMA} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {NgxEchartsModule} from 'ngx-echarts';
import {NgxGaugeModule} from 'ngx-gauge';
import {ElasticFilterAddComponent} from '../../shared/components/utm/filters/utm-elastic-filter/elastic-filter-add/elastic-filter-add.component';
import {UtmSharedModule} from '../../shared/utm-shared.module';
import {ChartViewComponent} from './components/viewer/chart-view/chart-view.component';
import {GoalViewComponent} from './components/viewer/goal-view/goal-view.component';
import {HeatMapLeafletViewComponent} from './components/viewer/heat-map-leaflet-view/heat-map-leaflet-view.component';
import {MapViewComponent} from './components/viewer/map-view/map-view.component';
import {MetricViewComponent} from './components/viewer/metric-view/metric-view.component';
import {TableViewComponent} from './components/viewer/table-view/table-view.component';
import {TextViewComponent} from './components/viewer/text-view/text-view.component';
import {UtmViewerComponent} from './components/viewer/utm-viewer/utm-viewer.component';

@NgModule({
  declarations: [
    ChartViewComponent,
    MetricViewComponent,
    GoalViewComponent,
    TableViewComponent,
    MapViewComponent,
    UtmViewerComponent,
    HeatMapLeafletViewComponent,
    TextViewComponent],
  exports: [
    ChartViewComponent,
    MetricViewComponent,
    GoalViewComponent,
    TableViewComponent,
    UtmViewerComponent
  ],
  imports: [
    CommonModule,
    NgbModule,
    UtmSharedModule,
    NgSelectModule,
    ReactiveFormsModule,
    NgxEchartsModule,
    FormsModule,
    NgxGaugeModule
  ],
  entryComponents: [
    ElasticFilterAddComponent,
    ChartViewComponent
  ],
  schemas: [CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA],
  providers: []
})
export class GraphicBuilderSharedModule {
}
