import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormArray, FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmMetricOptions} from '../../../../../shared/chart/types/charts/metric/utm-metric-options';
import {MetricAggregationType} from '../../../../../shared/chart/types/metric/metric-aggregation.type';
import {VisualizationType} from '../../../../../shared/chart/types/visualization.type';

import {IconSelectComponent} from '../../../../../shared/components/utm/util/icon-select/icon-select.component';
import {InputClassResolve} from '../../../../../shared/util/input-class-resolve';
import {MetricDataBehavior} from '../../shared/behaviors/metric-data.behavior';
import {extractMetricLabel} from '../../shared/functions/visualization-util';

@Component({
  selector: 'app-metric-properties-option',
  templateUrl: './metric-properties-option.component.html',
  styleUrls: ['./metric-properties-option.component.scss']
})
export class MetricPropertiesOptionComponent implements OnInit {
  @Input() visualization: VisualizationType;
  @Input() mode: string;
  @Output() metricOptions = new EventEmitter<UtmMetricOptions[]>();
  metrics: MetricAggregationType[] = [];
  formMetricOption: FormGroup;

  constructor(private modalService: NgbModal,
              private fb: FormBuilder,
              private inputClass: InputClassResolve,
              private metricDataBehavior: MetricDataBehavior) {
  }

  get options() {
    return this.formMetricOption.controls.options as FormArray;
  }

  ngOnInit() {
    this.initFormMetricOption();
    this.metricDataBehavior.$metric.subscribe(m => {
      this.metrics = m;
      this.extractDataMetric();
    });
    this.metricDataBehavior.$metricDeletedId.subscribe(id => {
      const indexOption = this.options.controls.findIndex(value =>
        value.get('metricId').value === id);
      if (indexOption > -1) {
        this.deleteOption(indexOption);
      }
    });

    if (this.mode === 'edit') {
      if (typeof this.visualization.chartConfig === 'string') {
        this.visualization.chartConfig = JSON.parse(this.visualization.chartConfig);
      }
      this.extractMetricOptionEdit();
    }
    this.formMetricOption.valueChanges.subscribe((value) => {
      this.metricOptions.emit(this.formMetricOption.get('options').value);
    });
  }

  extractMetricOptionEdit() {
    if (typeof this.visualization.chartConfig === 'string') {
      JSON.parse(this.visualization.chartConfig);
    }
    const config: UtmMetricOptions[] = this.visualization.chartConfig;
    for (const op of config) {
      const index = this.options.value.findIndex
      (value => {
        return Number(value.metricId) === Number(op.metricId);
      });
      if (index !== -1) {
        this.options.at(index).get('icon').setValue(op.icon);
        this.options.at(index).get('color').setValue(op.color);
        this.options.at(index).get('decimal').setValue(op.decimal);
      }
    }
  }

  extractDataMetric() {
    for (const m of this.metrics) {
      const indexOptions = this.options.controls.findIndex
      (value => value.get('metricId').value === m.id);
      if (indexOptions === -1) {
        this.addOption(m.id);
      }
    }
  }

  initFormMetricOption() {
    this.formMetricOption = this.fb.group({
      options: this.fb.array([]),
    });
  }

  addOption(metricId) {
    this.options.push(this.fb.group({
      metricId: [metricId, Validators.required],
      icon: [],
      color: ['#000000'],
      decimal: [0]
    }));
  }

  deleteOption(index) {
    this.options.removeAt(index);
  }

  selectIcon(index: number) {
    const modal = this.modalService.open(IconSelectComponent, {centered: true});
    modal.componentInstance.iconChange.subscribe(icon => {
      this.options.at(index).get('icon').setValue(icon);
    });
  }


  extractMetricLabel(index: number, visualization: VisualizationType): string {
    return extractMetricLabel(index, visualization);
  }
}
