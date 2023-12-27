import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup} from '@angular/forms';
import {Axis} from '../../../../../shared/chart/types/charts/chart-properties/axis/axis';
import {DataZoom} from '../../../../../shared/chart/types/charts/chart-properties/datazoom/data-zoom';
import {Grid} from '../../../../../shared/chart/types/charts/chart-properties/grid/grid';
import {Legend} from '../../../../../shared/chart/types/charts/chart-properties/legend/legend';
import {Toolbox} from '../../../../../shared/chart/types/charts/chart-properties/toolbox/toolbox';
import {Tooltip} from '../../../../../shared/chart/types/charts/chart-properties/tooltip/tooltip';
import {UtmLineBarOptionType} from '../../../../../shared/chart/types/charts/line/utm-line-bar-option.type';
import {VisualizationType} from '../../../../../shared/chart/types/visualization.type';
import {UTM_COLOR_THEME} from '../../../../../shared/constants/utm-color.const';
import {ChartTypeEnum} from '../../../../../shared/enums/chart-type.enum';

@Component({
  selector: 'app-line-bar-properties-option',
  templateUrl: './line-bar-properties-option.component.html',
  styleUrls: ['./line-bar-properties-option.component.scss']
})
export class LineBarPropertiesOptionComponent implements OnInit {
  @Input() visualization: VisualizationType;
  @Input() mode: string;
  @Output() lineOptions = new EventEmitter<UtmLineBarOptionType>();
  formLineBar: FormGroup;
  chartEnum = ChartTypeEnum;

  constructor(private fb: FormBuilder) {
  }

  ngOnInit() {
    this.initFormLine();
    this.formLineBar.valueChanges.subscribe(value => {
      if (value.seriesOption.length > 0) {
        this.lineOptions.emit(value);
      }
    });
  }

  initFormLine() {
    this.formLineBar = this.fb.group({
      color: [UTM_COLOR_THEME],
      legend: [new Legend(true)],
      toolbox: [new Toolbox(true)],
      yAxis: [this.visualization.chartType === this.chartEnum.BAR_HORIZONTAL_CHART ? new Axis('category') : new Axis('value')],
      xAxis: [this.visualization.chartType === this.chartEnum.BAR_HORIZONTAL_CHART ? new Axis('value') : new Axis('category')],
      grid: [new Grid()],
      tooltip: [new Tooltip('item')],
      dataZoom: [new DataZoom()],
      seriesOption: [],
    });
  }
}
