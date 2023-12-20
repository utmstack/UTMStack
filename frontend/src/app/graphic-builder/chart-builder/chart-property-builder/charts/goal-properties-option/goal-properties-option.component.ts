import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormArray, FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmGoalOption} from '../../../../../shared/chart/types/charts/goal/utm-goal-option';
import {MetricAggregationType} from '../../../../../shared/chart/types/metric/metric-aggregation.type';
import {VisualizationType} from '../../../../../shared/chart/types/visualization.type';
import {InputClassResolve} from '../../../../../shared/util/input-class-resolve';
import {MetricDataBehavior} from '../../shared/behaviors/metric-data.behavior';
import {extractMetricLabel} from '../../shared/functions/visualization-util';

@Component({
  selector: 'app-goal-properties-option',
  templateUrl: './goal-properties-option.component.html',
  styleUrls: ['./goal-properties-option.component.scss']
})
export class GoalPropertiesOptionComponent implements OnInit {
  @Input() visualization: VisualizationType;
  @Input() mode: string;
  @Output() goalOptions = new EventEmitter<UtmGoalOption[]>();
  metrics: MetricAggregationType[] = [];
  formGoalOption: FormGroup;
  cap = ['round', 'butt'];
  type = ['full', 'arch', 'semi'];

  constructor(private modalService: NgbModal,
              private fb: FormBuilder,
              private inputClass: InputClassResolve,
              private metricDataBehavior: MetricDataBehavior) {
  }

  get options() {
    return this.formGoalOption.controls.options as FormArray;
  }

  // extractGoalOptionEdit() {
  //   const config: UtmGoalOption[] = this.visualization.chartConfig;
  //   console.log(config);
  //   for (const op of config) {
  //     const index = this.options.controls.findIndex
  //     (value => value.get('goalId').value === op.goalId);
  //     console.log(index);
  //     this.options.at(index).get('icon').setValue(op.icon);
  //     this.options.at(index).get('color').setValue(op.color);
  //     this.options.at(index).get('decimal').setValue(op.decimal);
  //     console.log(this.options.value);
  //   }
  // }

  ngOnInit() {
    this.initFormGoalOption();

    this.metricDataBehavior.$metricDeletedId.subscribe(id => {
      const indexOption = this.options.controls.findIndex(value =>
        value.get('goalId').value === id);
      if (indexOption > -1) {
        this.deleteOption(indexOption);
      }
    });
    this.metricDataBehavior.$metric.subscribe(m => {
      this.metrics = m;
      this.extractDataGoal();
    });

    if (this.mode === 'edit') {
      if (typeof this.visualization.chartConfig === 'string') {
        this.visualization.chartConfig = JSON.parse(this.visualization.chartConfig);
      }
      // this.extractGoalOptionEdit();
    }

    this.formGoalOption.valueChanges.subscribe((value) => {
      this.goalOptions.emit(this.formGoalOption.get('options').value);
    });
    this.goalOptions.emit(this.formGoalOption.get('options').value);
  }

  extractDataGoal() {
    for (const m of this.metrics) {
      const indexOptions = this.options.controls.findIndex
      (value => {
        return value.get('metricId').value === m.id;
      });
      if (indexOptions === -1) {
        this.addOption(m, this.visualization);
      } else {
        this.resetLabelMetric();
      }
    }
  }

  initFormGoalOption() {
    this.formGoalOption = this.fb.group({
      options: this.fb.array([]),
    });
  }

  addOption(metric: MetricAggregationType, visualization: VisualizationType) {
    this.options.push(this.fb.group({
      metricId: [metric.id, Validators.required],
      value: [],
      max: [],
      min: [0],
      thick: [10],
      append: ['%'],
      animate: [true],
      cap: ['round'],
      type: ['arch'],
      thresholds: [{
        0: {color: 'red'},
        50: {color: 'orange'},
        75.5: {color: 'green'}
      }],
      foregroundColor: ['#43A047'],
      label: extractMetricLabel(metric.id, visualization),
      decimal: [2]
    }));
  }

  deleteOption(index) {
    this.options.removeAt(index);
  }


  extractMetricLabel(index: number, visualization: VisualizationType): string {
    return extractMetricLabel(index, visualization);
  }

  private resetLabelMetric() {
    for (const metric of this.metrics) {
      const indexOptions = this.options.controls.findIndex
      (value => value.get('metricId').value === metric.id);
      if (indexOptions > -1) {
        this.options.at(indexOptions).get('label').setValue(extractMetricLabel(metric.id, this.visualization));
      }
    }
  }
}
