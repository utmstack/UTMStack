import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup} from '@angular/forms';
import {Grid} from '../../../../../shared/chart/types/charts/chart-properties/grid/grid';
import {Legend} from '../../../../../shared/chart/types/charts/chart-properties/legend/legend';
import {Toolbox} from '../../../../../shared/chart/types/charts/chart-properties/toolbox/toolbox';
import {UtmPieOptionType} from '../../../../../shared/chart/types/charts/pie/utm-pie-option.type';
import {VisualizationType} from '../../../../../shared/chart/types/visualization.type';
import {UTM_COLOR_THEME} from '../../../../../shared/constants/utm-color.const';
import {ChartTypeEnum} from '../../../../../shared/enums/chart-type.enum';

@Component({
  selector: 'app-pie-properties-options',
  templateUrl: './pie-properties-options.component.html',
  styleUrls: ['./pie-properties-options.component.scss']
})
export class PiePropertiesOptionsComponent implements OnInit {
  @Output() pieOptions = new EventEmitter<UtmPieOptionType>();
  @Input() visualization: VisualizationType;
  @Input() mode: string;
  chartConfig: UtmPieOptionType;
  formPieOptions: FormGroup;
  advanced = false;
  property: string;
  chartTypeEnum = ChartTypeEnum;


  constructor(private fb: FormBuilder) {
  }

  ngOnInit() {
    this.chartConfig = this.visualization.chartConfig;
    this.initFormPie();
    this.formPieOptions.valueChanges.subscribe(value => {
      this.pieOptions.emit(value);
    });
    if (this.mode === 'edit') {
      this.chartConfig = (typeof this.visualization.chartConfig === 'string') ?
        JSON.parse(this.visualization.chartConfig) :
        this.visualization.chartConfig;
      this.formPieOptions.patchValue(this.chartConfig);
    }
    this.pieOptions.emit(this.formPieOptions.value);
  }

  initFormPie() {
    this.formPieOptions = this.fb.group({
      pieType: ['donut'],
      legend: [new Legend(true, [])],
      color: [UTM_COLOR_THEME],
      toolbox: [new Toolbox()],
      grid: [new Grid()]
    });
  }

}
