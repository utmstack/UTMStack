import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup} from '@angular/forms';
import {Toolbox} from '../../../../../shared/chart/types/charts/chart-properties/toolbox/toolbox';
import {Tooltip} from '../../../../../shared/chart/types/charts/chart-properties/tooltip/tooltip';
import {UtmTagCloudOptionType} from '../../../../../shared/chart/types/charts/tag-cloud/utm-tag-cloud-option.type';
import {VisualizationType} from '../../../../../shared/chart/types/visualization.type';
import {UTM_COLOR_THEME} from '../../../../../shared/constants/utm-color.const';

@Component({
  selector: 'app-tag-cloud-properties-option',
  templateUrl: './tag-cloud-properties-option.component.html',
  styleUrls: ['./tag-cloud-properties-option.component.scss']
})
export class TagCloudPropertiesOptionComponent implements OnInit {
  @Output() tagCloudOptions = new EventEmitter<UtmTagCloudOptionType>();
  @Input() visualization: VisualizationType;
  @Input() mode: string;
  formTagCloud: FormGroup;
  tagShapes = ['circle', 'cardioid', 'diamond', 'triangle-forward', 'triangle', 'star'];
  options: UtmTagCloudOptionType = {
    toolbox: new Toolbox(),
    tooltip: new Tooltip('item'),
    series: null
  };

  constructor(private fb: FormBuilder) {
  }

  ngOnInit() {
    this.initFormTagCloud();
    this.formTagCloud.valueChanges.subscribe(value => {
      this.options.series = [value];
      this.tagCloudOptions.emit(this.options);
    });
    this.options.series = [this.formTagCloud.value];
    this.tagCloudOptions.emit(this.options);
  }

  initFormTagCloud() {
    this.formTagCloud = this.fb.group({
      size: [['100%', '100%']],
      textRotation: [[0, 45, 90, -45]],
      textPadding: [0],
      shape: ['circle'],
      color: [UTM_COLOR_THEME],
      type: ['wordCloud'],
      autoSize: this.fb.group({
        enable: [true],
        minSize: [14]
      }),
      data: [[]]
    });
  }


  onColorOptionChange($event: string[]) {
    this.formTagCloud.get('color').setValue($event);
    this.tagCloudOptions.emit(this.options);
  }

  onToolboxOptionChange($event: Toolbox) {
    this.options.toolbox = $event;
    this.tagCloudOptions.emit(this.options);
  }

}
