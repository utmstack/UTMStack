import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup} from '@angular/forms';
import {Toolbox} from '../../../../../shared/chart/types/charts/chart-properties/toolbox/toolbox';
import {ChartTypeEnum} from '../../../../../shared/enums/chart-type.enum';
import {
  TOOLBOX_MAGIC_TYPE,
  TOOLBOX_POSITION_LEFT,
  TOOLBOX_POSITION_ORIENTATION,
  TOOLBOX_POSITION_TOP
} from '../../../../shared/const/chart/toolbox-properties.const';

@Component({
  selector: 'app-chart-toolbox-option',
  templateUrl: './chart-toolbox-option.component.html',
  styleUrls: ['./chart-toolbox-option.component.scss']
})
export class ChartToolboxOptionComponent implements OnInit {
  @Input() chartType: ChartTypeEnum;
  @Output() toolboxOptionChange = new EventEmitter<Toolbox>();
  formToolboxOption: FormGroup;
  chartEnum = ChartTypeEnum;
  viewToolbox = false;
  toolboxPositionsTop = TOOLBOX_POSITION_TOP;
  toolboxPositionsLeft = TOOLBOX_POSITION_LEFT;
  toolboxOrientations = TOOLBOX_POSITION_ORIENTATION;
  toolboxMagicType = TOOLBOX_MAGIC_TYPE;

  constructor(private fb: FormBuilder) {
  }

  ngOnInit() {
    this.initFormToolbox();
    this.formToolboxOption.valueChanges.subscribe(value => {
      this.toolboxOptionChange.emit(this.formToolboxOption.value);
    });
    this.toolboxOptionChange.emit(this.formToolboxOption.value);
    if (this.chartType === this.chartEnum.PIE_CHART ||
      this.chartType === this.chartEnum.GAUGE_CHART ||
      this.chartType === this.chartEnum.TAG_CLOUD_CHART
    ) {
      this.formToolboxOption.get(['feature', 'magicType', 'show']).setValue(false);
      this.formToolboxOption.get(['feature', 'dataZoom', 'show']).setValue(false);
      this.formToolboxOption.get(['feature', 'mark', 'show']).setValue(false);
    }
  }

  initFormToolbox() {
    this.formToolboxOption = this.fb.group({
      show: [true],
      feature: this.fb.group(
        {
          saveAsImage: this.fb.group({
            show: [true],
            type: ['png'],
            name: ['utm-chart'],
            title: ['Save as image']
          }),
          restore: this.fb.group({
            show: [true],
            title: ['Restore'],
          }),
          dataView: this.fb.group({
            show: [true],
            title: ['Data view'],
            readOnly: [true],
            lang: [['Data view', 'Close', 'Refresh']],
            backgroundColor: [],
            textareaColor: [],
            textareaBorderColor: [],
            textColor: [],
            buttonColor: ['#0277bd'],
            buttonTextColor: ['#fff']
          }),
          dataZoom: this.fb.group({
            show: [true],
            title: [{
              zoom: ['Zoom'],
              back: ['Step back']
            }]
          }),
          magicType: this.fb.group({
            show: [true],
            type: [],
            title: this.fb.group({
              line: 'Line',
              bar: 'Bar',
              stack: 'Stack',
              tiled: 'Tiled'
            })
          }),
          brush: this.fb.group({
            type: ['rect']
          }),
          mark: this.fb.group({
            show: [true]
          })
        }
      ),
      orient: ['horizontal'],
      itemSize: [14],
      showTitle: [],
      left: ['right'],
      top: ['top'],
      width: [],
      height: [],
      iconStyle: []
    });
  }

  viewToolboxProperties() {
    this.viewToolbox = this.viewToolbox ? false : true;
  }
}
