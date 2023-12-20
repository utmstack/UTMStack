import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup} from '@angular/forms';
import {Legend} from '../../../../../shared/chart/types/charts/chart-properties/legend/legend';
import {
  LEGEND_ICONS,
  LEGEND_ORIENTATION,
  LEGEND_POSITION_LEFT,
  LEGEND_POSITION_TOP
} from '../../../../shared/const/chart/legend-properties.const';

@Component({
  selector: 'app-chart-legend-option',
  templateUrl: './chart-legend-option.component.html',
  styleUrls: ['./chart-legend-option.component.scss']
})
export class ChartLegendOptionComponent implements OnInit {
  @Input() legend: Legend;
  @Input() mode: string;
  @Output() legendOptionChange = new EventEmitter<Legend>();
  formLegendOption: FormGroup;
  legendPositionsTop = LEGEND_POSITION_TOP;
  legendPositionsLeft = LEGEND_POSITION_LEFT;
  legendOrientations = LEGEND_ORIENTATION;
  legendIcons = LEGEND_ICONS;
  viewLegend = false;
  useIcon = false;

  constructor(private fb: FormBuilder) {
  }


  ngOnInit() {
    this.initFormPie();
    if (this.mode === 'edit') {
      this.formLegendOption.patchValue(this.legend, {emitEvent: true});
    }
    this.formLegendOption.valueChanges.subscribe(value => {
      this.legendOptionChange.emit(this.formLegendOption.value);
    });
    this.legendOptionChange.emit(this.formLegendOption.value);
  }

  initFormPie() {
    this.formLegendOption = this.fb.group({
      show: [true],
      type: ['scroll'],
      top: ['bottom'],
      left: ['center'],
      orient: ['horizontal'],
      itemHeight: [8],
      itemWidth: [8],
      icon: ['roundRect'],
      extraCssText: ['z-index:100']
    });
  }

  viewLegendProperties() {
    this.viewLegend = this.viewLegend ? false : true;
  }

  useCustomIcon($event: boolean) {
    this.formLegendOption.get('icon').setValue('');
    this.useIcon = $event;
  }


}
