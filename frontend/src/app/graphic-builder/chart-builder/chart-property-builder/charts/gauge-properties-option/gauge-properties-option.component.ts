import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormArray, FormBuilder, FormGroup} from '@angular/forms';
import {Grid} from '../../../../../shared/chart/types/charts/chart-properties/grid/grid';
import {Toolbox} from '../../../../../shared/chart/types/charts/chart-properties/toolbox/toolbox';
import {Tooltip} from '../../../../../shared/chart/types/charts/chart-properties/tooltip/tooltip';
import {UtmGaugeOptionType} from '../../../../../shared/chart/types/charts/gauge/utm-gauge-option.type';
import {VisualizationType} from '../../../../../shared/chart/types/visualization.type';
import {ChartTypeEnum} from '../../../../../shared/enums/chart-type.enum';
import {AXIS_LINE_STYLE} from '../../../../shared/const/chart/axis-properties.const';

@Component({
  selector: 'app-gauge-properties-option',
  templateUrl: './gauge-properties-option.component.html',
  styleUrls: ['./gauge-properties-option.component.scss']
})
export class GaugePropertiesOptionComponent implements OnInit {
  @Output() gaugeOptions = new EventEmitter<UtmGaugeOptionType>();
  @Input() visualization: VisualizationType;
  @Input() mode: string;
  lineStyle = AXIS_LINE_STYLE;
  chartEnum = ChartTypeEnum;
  options: UtmGaugeOptionType = {
    toolbox: new Toolbox(),
    grid: new Grid(),
    tooltip: new Tooltip('axis'),
    serie: []
  };
  chartConfig: UtmGaugeOptionType;
  formGaugeSerie: FormGroup;
  viewGauge = false;
  viewAxisTick = false;
  viewPointer = false;
  viewDetail = false;
  viewAxisLine = false;

  constructor(private fb: FormBuilder) {
  }

  get series() {
    return this.formGaugeSerie.controls.series as FormArray;
  }

  ngOnInit() {
    this.initGaugeForm();
    this.formGaugeSerie.valueChanges.subscribe((val) => {
      this.options.serie = this.formGaugeSerie.get('series').value;
      this.gaugeOptions.emit(this.options);
    });
    this.addSerie();
  }

  initGaugeForm() {
    this.formGaugeSerie = this.fb.group({
      series: this.fb.array([])
    });
  }

  addSerie() {
    this.series.push(this.fb.group({
      type: ['gauge'],
      startAngle: [180],
      endAngle: [0],
      min: [0],
      max: [100],
      center: [['50%', '70%']],
      radius: ['100%'],
      splitNumber: [10],
      axisTick: this.fb.group({
        show: true,
        splitNumber: 10,
        length: 10,
        lineStyle: this.fb.group({
          color: '#eee', width: 0.5, type: 'solid'
        })
      }),
      pointer: this.fb.group({
        show: true,
        width: 6,
        length: '80%'
      }),
      detail: this.fb.group({
        fontWeight: 'bolder',
        formatter: '{value}%',
        offsetCenter: [[0, 30]]
      }),
      axisLine: this.fb.group({
        show: [true],
        lineStyle: this.fb.group({
          color:
            [[
              [.3, '#388E3C'],
              [.7, '#FB8C00'],
              [1, '#E53935']
            ]],
          width: 12
        })
      }),
      title: {
        offsetCenter: [0, -115],
        color: 'auto'
      },
      splitLine: {
        length: [15],
        lineStyle: {
          color: 'auto'
        }
      }
    }));
  }

  deleteSeries(index) {
    const id = this.series.at(index).get('id').value;
    this.series.removeAt(index);
  }

  onGridOptionChange($event: Grid) {
    this.options.grid = $event;
    this.gaugeOptions.emit(this.options);
  }

  onToolboxOptionChange($event: Toolbox) {
    this.options.toolbox = $event;
    this.gaugeOptions.emit(this.options);
  }

  viewGaugeProperties() {
    this.viewGauge = this.viewGauge ? false : true;
  }

  viewAxisTickProperties() {
    this.viewAxisTick = this.viewAxisTick ? false : true;
  }

  viewPointerProperties() {
    this.viewPointer = this.viewPointer ? false : true;
  }

  viewDetailProperties() {
    this.viewDetail = this.viewDetail ? false : true;
  }

  setOffset($event, position: string, index: number) {
    const offset: number[] = this.series.at(index).get('detail').get('offsetCenter').value;
    position === 'left' ? offset[0] = Number($event.target.value) : offset[1] = Number($event.target.value);
    this.series.at(index).get('detail').get('offsetCenter').setValue(offset);
  }

  viewAxisLineProperties() {
    this.viewAxisLine = this.viewAxisLine ? false : true;
  }
}
