import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup} from '@angular/forms';
import {Axis} from '../../../../../shared/chart/types/charts/chart-properties/axis/axis';
import {ChartTypeEnum} from '../../../../../shared/enums/chart-type.enum';
import {AXIS_LINE_STYLE, AXIS_TYPE} from '../../../../shared/const/chart/axis-properties.const';

@Component({
  selector: 'app-chart-axis-option',
  templateUrl: './chart-axis-option.component.html',
  styleUrls: ['./chart-axis-option.component.scss']
})
export class ChartAxisOptionComponent implements OnInit {
  /**
   * Type of axis
   */
  @Input() axis: 'yAxis' | 'xAxis';
  /**
   * Determine if only accept category as axis type
   */
  @Input() onlyCategory: boolean;

  @Input() chartType: ChartTypeEnum;
  /**
   * Fire event when form axis change
   */
  @Output() axisOptionChange = new EventEmitter<Axis>();
  axisType = AXIS_TYPE;
  axisLineStyle = AXIS_LINE_STYLE;
  formAxisOption: FormGroup;
  viewAxis = false;
  chartEnum = ChartTypeEnum;

  constructor(private fb: FormBuilder) {
  }

  ngOnInit() {
    this.initFormPie();
    this.formAxisOption.valueChanges.subscribe(() => {
      this.axisOptionChange.emit(this.formAxisOption.value);
    });
    this.axisOptionChange.emit(this.formAxisOption.value);
  }

  initFormPie() {
    this.formAxisOption = this.fb.group({
      name: [],
      type: [this.resolveChartType()],
      data: [],
      axisLabel: this.fb.group({
        color: ['#333'],
        formatter: ['{value}']
      }),
      axisLine: this.fb.group({
        lineStyle: this.fb.group({
          color: ['#999']
        })
      }),
      splitLine: this.fb.group({
        show: [true],
        lineStyle: this.fb.group({
          color: ['#eee'],
          type: ['dashed']
        })
      })
    });
  }

  resolveChartType(): 'value' | 'category' {
    if (this.onlyCategory) {
      return 'category';
    } else if (this.chartType !== this.chartEnum.BAR_HORIZONTAL_CHART) {
      return this.axis === 'xAxis' ? 'category' : 'value';
    } else if (this.chartType === this.chartEnum.BAR_HORIZONTAL_CHART) {
      return this.axis === 'xAxis' ? 'value' : 'category';
    }
  }

  viewAxisProperties() {
    this.viewAxis = this.viewAxis ? false : true;
  }
}
