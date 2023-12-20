import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup} from '@angular/forms';
import {Axis} from '../../../../../shared/chart/types/charts/chart-properties/axis/axis';
import {DataZoom} from '../../../../../shared/chart/types/charts/chart-properties/datazoom/data-zoom';
import {Grid} from '../../../../../shared/chart/types/charts/chart-properties/grid/grid';
import {Legend} from '../../../../../shared/chart/types/charts/chart-properties/legend/legend';
import {Toolbox} from '../../../../../shared/chart/types/charts/chart-properties/toolbox/toolbox';
import {Tooltip} from '../../../../../shared/chart/types/charts/chart-properties/tooltip/tooltip';
import {VisualMap} from '../../../../../shared/chart/types/charts/chart-properties/visualmap/visual-map';
import {HeatMapPropertiesType} from '../../../../../shared/chart/types/charts/heatmap/heat-map-properties.type';
import {UTM_COLOR_THEME} from '../../../../../shared/constants/utm-color.const';

@Component({
  selector: 'app-heat-map-properties-option',
  templateUrl: './heat-map-properties-option.component.html',
  styleUrls: ['./heat-map-properties-option.component.scss']
})
export class HeatMapPropertiesOptionComponent implements OnInit {
  @Output() heatMapOptionChange = new EventEmitter<HeatMapPropertiesType>();
  formHeatMapOption: FormGroup;

  constructor(private fb: FormBuilder) {
  }

  ngOnInit() {
    this.initFormHeatMap();
    this.formHeatMapOption.valueChanges.subscribe(value => this.heatMapOptionChange.emit(value));
    this.heatMapOptionChange.emit(this.formHeatMapOption.value);
  }

  initFormHeatMap() {
    this.formHeatMapOption = this.fb.group({
      legend: [new Legend(true)],
      toolbox: [new Toolbox()],
      xAxis: [new Axis('category')],
      yAxis: [new Axis('category')],
      series: [],
      color: [UTM_COLOR_THEME],
      grid: [new Grid()],
      tooltip: [new Tooltip('item')],
      dataZoom: [new DataZoom()],
      visualMap: [new VisualMap()],
    });
  }

  viewHeatMapOptionProperties() {
  }
}
