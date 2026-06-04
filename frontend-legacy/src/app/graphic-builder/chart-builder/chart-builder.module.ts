import {CommonModule} from '@angular/common';
import {CUSTOM_ELEMENTS_SCHEMA, NgModule, NO_ERRORS_SCHEMA} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {RouterModule} from '@angular/router';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {NgxEchartsModule} from 'ngx-echarts';
import {NgxSpinnerModule} from 'ngx-spinner';
import {InputClassResolve} from '../../shared/util/input-class-resolve';
import {UtmSharedModule} from '../../shared/utm-shared.module';
import {GraphicBuilderSharedModule} from '../shared/graphic-builder-shared.module';
import {VisualizationSharedModule} from '../visualization/visualization-shared.module';
import {ChartBuilderRoutingModule} from './chart-builder-routing.module';
import {ChartBuilderComponent} from './chart-builder.component';
import {BucketAggregationComponent} from './chart-property-builder/bucket-aggregation/bucket-aggregation.component';
// tslint:disable-next-line:max-line-length
import {BucketDateHistogramComponent} from './chart-property-builder/bucket-aggregation/bucket-date-histogram/bucket-date-histogram.component';
import {BucketRangeComponent} from './chart-property-builder/bucket-aggregation/bucket-range/bucket-range.component';
// tslint:disable-next-line:max-line-length
import {BucketSignificantTermsComponent} from './chart-property-builder/bucket-aggregation/bucket-significant-terms/bucket-significant-terms.component';
import {BucketTermsComponent} from './chart-property-builder/bucket-aggregation/bucket-terms/bucket-terms.component';
import {ChartActionComponent} from './chart-property-builder/chart-action/chart-action.component';
import {ChartAxisOptionComponent} from './chart-property-builder/chart-options/chart-axis-option/chart-axis-option.component';
import {ChartColorOptionComponent} from './chart-property-builder/chart-options/chart-color-option/chart-color-option.component';
import {ChartDataZoomComponent} from './chart-property-builder/chart-options/chart-data-zoom/chart-data-zoom.component';
import {ChartGridOptionComponent} from './chart-property-builder/chart-options/chart-grid-option/chart-grid-option.component';
import {ChartLegendOptionComponent} from './chart-property-builder/chart-options/chart-legend-option/chart-legend-option.component';
import {ChartMapTilesOptionComponent} from './chart-property-builder/chart-options/chart-map-tiles-option/chart-map-tiles-option.component';
import {ChartMarkOptionComponent} from './chart-property-builder/chart-options/chart-mark-option/chart-mark-option.component';
// tslint:disable-next-line:max-line-length
import {ChartSeriesLineBarOptionComponent} from './chart-property-builder/chart-options/chart-series-line-bar-option/chart-series-line-bar-option.component';
import {ChartToolboxOptionComponent} from './chart-property-builder/chart-options/chart-toolbox-option/chart-toolbox-option.component';
// tslint:disable-next-line:max-line-length
import {ChartVisualMapOptionComponent} from './chart-property-builder/chart-options/chart-visual-map-option/chart-visual-map-option.component';
import {MapLeafletOptionsComponent} from './chart-property-builder/chart-options/map-leaflet-options/map-leaflet-options.component';
import {ChartPropertyHeaderComponent} from './chart-property-builder/chart-property-header/chart-property-header.component';
import {GaugePropertiesOptionComponent} from './chart-property-builder/charts/gauge-properties-option/gauge-properties-option.component';
// tslint:disable-next-line:max-line-length
import {GaugeRangeColorComponent} from './chart-property-builder/charts/gauge-properties-option/shared/gauge-range-color/gauge-range-color.component';
import {GoalPropertiesOptionComponent} from './chart-property-builder/charts/goal-properties-option/goal-properties-option.component';
// tslint:disable-next-line:max-line-length
import {HeatMapPropertiesOptionComponent} from './chart-property-builder/charts/heat-map-properties-option/heat-map-properties-option.component';
// tslint:disable-next-line:max-line-length
import {LineBarPropertiesOptionComponent} from './chart-property-builder/charts/line-bar-properties-option/line-bar-properties-option.component';
import {MetricPropertiesOptionComponent} from './chart-property-builder/charts/metric-properties-option/metric-properties-option.component';
import {PiePropertiesOptionsComponent} from './chart-property-builder/charts/pie-properties-options/pie-properties-options.component';
// tslint:disable-next-line:max-line-length
import {ScatterMapPropertiesOptionComponent} from './chart-property-builder/charts/scatter-map-properties-option/scatter-map-properties-option.component';
import {TablePropertiesOptionComponent} from './chart-property-builder/charts/table-properties-option/table-properties-option.component';
// tslint:disable-next-line:max-line-length
import {TagCloudPropertiesOptionComponent} from './chart-property-builder/charts/tag-cloud-properties-option/tag-cloud-properties-option.component';
import {ListColumnsComponent} from './chart-property-builder/list-columns/list-columns.component';
import {MetricAggregationComponent} from './chart-property-builder/metric-agregation/metric-aggregation.component';

@NgModule({
  declarations: [
    ChartBuilderComponent,
    ChartPropertyHeaderComponent,
    PiePropertiesOptionsComponent,
    ChartLegendOptionComponent,
    ChartColorOptionComponent,
    ChartToolboxOptionComponent,
    ChartGridOptionComponent,
    ChartAxisOptionComponent,
    ChartDataZoomComponent,
    BucketAggregationComponent,
    BucketTermsComponent,
    BucketDateHistogramComponent,
    MetricAggregationComponent,
    MetricPropertiesOptionComponent,
    BucketSignificantTermsComponent,
    BucketRangeComponent,
    GaugePropertiesOptionComponent,
    GaugeRangeColorComponent,
    GoalPropertiesOptionComponent,
    TablePropertiesOptionComponent,
    TagCloudPropertiesOptionComponent,
    LineBarPropertiesOptionComponent,
    ChartMarkOptionComponent,
    ChartSeriesLineBarOptionComponent,
    ScatterMapPropertiesOptionComponent,
    ChartMapTilesOptionComponent,
    MapLeafletOptionsComponent,
    ChartVisualMapOptionComponent,
    HeatMapPropertiesOptionComponent,
    ChartActionComponent,
    ListColumnsComponent,
  ],
  entryComponents: [],
  imports: [
    CommonModule,
    ChartBuilderRoutingModule,
    NgxEchartsModule,
    RouterModule,
    UtmSharedModule,
    NgbModule,
    FormsModule,
    NgSelectModule,
    GraphicBuilderSharedModule,
    NgxSpinnerModule,
    ReactiveFormsModule,
    VisualizationSharedModule
  ],
  schemas: [CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA],
  exports: [],
  providers: [InputClassResolve]
})
export class ChartBuilderModule {
}
