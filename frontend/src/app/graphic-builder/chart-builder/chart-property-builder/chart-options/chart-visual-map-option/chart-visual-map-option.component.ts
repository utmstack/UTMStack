import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup} from '@angular/forms';
import {VisualMap} from '../../../../../shared/chart/types/charts/chart-properties/visualmap/visual-map';
import {
  VISUAL_MAP_ORIENT,
  VISUAL_MAP_POSITION_LEFT,
  VISUAL_MAP_POSITION_TOP
} from '../../../../shared/const/chart/visual-map-properties.const';

@Component({
  selector: 'app-chart-visual-map-option',
  templateUrl: './chart-visual-map-option.component.html',
  styleUrls: ['./chart-visual-map-option.component.scss']
})
export class ChartVisualMapOptionComponent implements OnInit {
  @Output() visualMapOptionChange = new EventEmitter<VisualMap>();
  formVisualMapOption: FormGroup;
  viewVisualMap = false;
  visualMapTypes = ['continuous', 'piecewise'];
  visualMapLeftPositions = VISUAL_MAP_POSITION_LEFT;
  visualMapTopPositions = VISUAL_MAP_POSITION_TOP;
  visualMapOrient = VISUAL_MAP_ORIENT;

  constructor(private fb: FormBuilder) {
  }

  ngOnInit() {
    this.initFormVisualMap();
    this.formVisualMapOption.valueChanges.subscribe(value => this.visualMapOptionChange.emit(value));
  }

  initFormVisualMap() {
    this.formVisualMapOption = this.fb.group({
      type: 'continuous',
      show: true,
      calculable: true,
      realtime: true,
      inRange: this.fb.group({
        color: [['#f5994e', '#c05050']]
      }),
      orient: 'horizontal',
      left: 'center',
      top: 'bottom',
      bottom: '15%',
      min: [],
      max: [],
      colorStart: ['#f5994e'],
      colorEnd: ['#c05050'],
    });
  }

  viewVisualMapProperties() {
    this.viewVisualMap = this.viewVisualMap ? false : true;
  }

  setColorStart($event: string) {
    this.formVisualMapOption.get('colorStart').setValue($event);
    this.formVisualMapOption.get(['inRange', 'color']).setValue(
      [this.formVisualMapOption.get('colorStart').value, this.formVisualMapOption.get('colorEnd').value]
    );
  }

  setColorEnd($event: string) {
    this.formVisualMapOption.get('colorEnd').setValue($event);
    this.formVisualMapOption.get(['inRange', 'color']).setValue(
      [this.formVisualMapOption.get('colorStart').value, this.formVisualMapOption.get('colorEnd').value]
    );
  }
}
