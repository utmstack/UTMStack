import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormArray, FormBuilder, FormGroup, Validators} from '@angular/forms';
import {Observable} from 'rxjs';
import {MetricAggregationType} from '../../../../shared/chart/types/metric/metric-aggregation.type';
import {VisualizationType} from '../../../../shared/chart/types/visualization.type';
import {ElasticDataTypesEnum} from '../../../../shared/enums/elastic-data-types.enum';
import {FieldDataService} from '../../../../shared/services/elasticsearch/field-data.service';
import {ElasticSearchFieldInfoType} from '../../../../shared/types/elasticsearch/elastic-search-field-info.type';
import {InputClassResolve} from '../../../../shared/util/input-class-resolve';
import {AGGREGATIONS} from '../../../shared/const/aggregation.const';
import {filterFieldDataType} from '../../../shared/util/elasticsearch/elastic-field.util';
import {MetricDataBehavior} from '../shared/behaviors/metric-data.behavior';
import {getErrorCountFormArray} from '../shared/functions/form-error-count';
import {filterFieldAgg} from '../shared/functions/util-field';

@Component({
  selector: 'app-metric-aggregation',
  templateUrl: './metric-aggregation.component.html',
  styleUrls: ['./metric-aggregation.component.scss']
})
export class MetricAggregationComponent implements OnInit {
  @Input() allowMultiMetric: boolean;
  @Input() visualization: VisualizationType;
  @Input() mode: string;
  @Output() metricDefChange = new EventEmitter<MetricAggregationType[]>();
  @Output() metricErrors = new EventEmitter<boolean>();
  errors: string[] = [];
  display: 'data' | 'options';
  aggregations = AGGREGATIONS;
  formMetric: FormGroup;
  fields: ElasticSearchFieldInfoType[] = [];
  allFields: ElasticSearchFieldInfoType[] = [];
  viewAgg = 0;
  showAdvanced = -1;
  metricId = 1;

  constructor(private fb: FormBuilder,
              public inputClass: InputClassResolve,
              private fieldDataBehavior: FieldDataService,
              private metricDataBehavior: MetricDataBehavior) {
  }

  get metrics() {
    return this.formMetric.controls.metrics as FormArray;
  }

  ngOnInit() {
    this.fieldDataBehavior.$fields.subscribe(data => {
      this.fields = filterFieldAgg(data);
      this.allFields = filterFieldAgg(data);
    });
    this.initFormMetric();
    if (this.mode === 'edit') {
      this.addMetricEdit(this.visualization.aggregationType.metrics).subscribe(() => {
        this.metricDataBehavior.$metric.next(this.formMetric.get('metrics').value);
        this.metricDefChange.emit(this.formMetric.get('metrics').value);
      });
      this.metricId = this.visualization.aggregationType.metrics.length + 1;
    } else {
      this.addMetric();
      this.metricDataBehavior.$metric.next(this.formMetric.get('metrics').value);
    }
    this.formMetric.valueChanges.subscribe(() => {
      this.metricDefChange.emit(this.formMetric.get('metrics').value);
      this.metricDataBehavior.$metric.next(this.formMetric.get('metrics').value);
      this.metricErrors.emit(this.formMetric.valid);
    });
  }

  initFormMetric() {
    this.formMetric = this.fb.group({
      metrics: this.fb.array([], Validators.required),
    });
  }

  addMetric() {
    this.metrics.push(this.fb.group({
      id: [this.metricId, Validators.required],
      aggregation: ['COUNT', Validators.required],
      field: [''],
      customLabel: [''],
    }));
    this.viewAgg = this.metrics.length - 1;
    this.metricId += 1;
    this.metricDataBehavior.$metric.next(this.formMetric.get('metrics').value);
    this.metricDefChange.emit(this.formMetric.get('metrics').value);
  }

  addMetricEdit(metric: MetricAggregationType[]): Observable<string> {
    return new Observable<string>(subscriber => {
      for (const m of metric) {
        this.metrics.push(this.fb.group({
          id: [m.id, Validators.required],
          aggregation: [m.aggregation, Validators.required],
          field: [m.field],
          customLabel: [m.customLabel],
        }));
      }
      subscriber.next('finish');
    });
  }

  deleteMetric(index) {
    const id = this.metrics.at(index).get('id').value;
    this.metricDataBehavior.$metricDeletedId.next(id);
    this.metrics.removeAt(index);
  }

  viewAggregationDetail(index: number) {
    this.viewAgg = this.viewAgg === index ? -1 : index;
  }

  showAdvancedJson(index: number) {
    if (this.showAdvanced !== index) {
      this.showAdvanced = index;
    } else {
      this.showAdvanced = -1;
    }
  }

  countErrors(): number {
    return getErrorCountFormArray(this.metrics);
  }


  changeFieldsAllowed($event, index: number) {
    if ($event.value !== 'COUNT') {
      this.addValidatorOnMetricIndex(index);
      this.metrics.at(index).get('field').markAsDirty();
      this.metrics.at(index).get('field').markAsTouched();
      this.metrics.at(index).get('field').markAsPristine();
    } else if ($event.value === 'COUNT') {
      this.removeValidatorOnMetricIndex(index);
    }

    if ($event.value !== 'UNIQUE_COUNT') {
      filterFieldDataType(this.allFields, [ElasticDataTypesEnum.LONG, ElasticDataTypesEnum.FLOAT]).subscribe(values => {
        this.fields = values;
      });
    } else {
      this.fields = this.allFields;
    }
  }

  addValidatorOnMetricIndex(index: number) {
    this.metrics.at(index).get('field').setValidators(Validators.required);
    this.metrics.at(index).get('field').updateValueAndValidity();
    this.metrics.updateValueAndValidity();
    this.formMetric.updateValueAndValidity();
  }

  removeValidatorOnMetricIndex(index: number) {
    this.metrics.at(index).get('field').setValidators(null);
    this.metrics.at(index).get('field').updateValueAndValidity();
    this.metrics.updateValueAndValidity();
    this.formMetric.updateValueAndValidity();
  }


  // verifyForm(type: string): Observable<string[]> {
  //   return new Observable<string[]>(subscriber => {
  //     for (const metric of this.metrics.controls) {
  //       this.errors = [];
  //       // @ts-ignore
  //       Object.keys(metric['controls']).forEach(control => {
  //         if (!metric['controls'][control].valid) {
  //           Object.keys(metric['controls'][control].errors).forEach((error) => {
  //             if (error === 'required') {
  //               this.errors.push('Error in ' + type + ', ' + control + ' is required');
  //             }
  //             if (error === 'pattern' || error === 'min' || error === 'max') {
  //               this.errors.push(control + ' is invalid');
  //             }
  //           });
  //         }
  //       });
  //     }
  //     subscriber.next(this.errors);
  //   });
  // }
}
