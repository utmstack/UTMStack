import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {BucketTermsType} from '../../../../../shared/chart/types/metric/bucket-terms.type';
import {MetricAggregationType} from '../../../../../shared/chart/types/metric/metric-aggregation.type';
import {SortByType} from '../../../../../shared/types/sort-by.type';
import {InputClassResolve} from '../../../../../shared/util/input-class-resolve';
import {MetricDataBehavior} from '../../shared/behaviors/metric-data.behavior';
import {getErrorCountForm} from '../../shared/functions/form-error-count';

@Component({
  selector: 'app-bucket-terms',
  templateUrl: './bucket-terms.component.html',
  styleUrls: ['./bucket-terms.component.scss']
})
export class BucketTermsComponent implements OnInit {
  @Input() mode: string;
  @Input() terms: BucketTermsType;
  @Output() termBucketChange = new EventEmitter<BucketTermsType>();
  @Output() errorsCount = new EventEmitter<number>();
  fields: SortByType[] = [];
  sort: { name: string, value: boolean }[] = [
    {name: 'Ascending', value: true},
    {name: 'Descending', value: false}];
  formBucketTerms: FormGroup;

  constructor(private metricDataBehavior: MetricDataBehavior,
              private fb: FormBuilder,
              public inputClass: InputClassResolve) {
    this.metricDataBehavior.$metric.subscribe(m => {
      this.loadFieldSort(m);
    });
  }

  ngOnInit() {
    this.initFormBucketTerms();
    if (this.mode === 'edit') {
      this.formBucketTerms.patchValue(this.terms);
    }
    this.termBucketChange.emit(this.formBucketTerms.value);
    this.formBucketTerms.valueChanges.subscribe(() => {
      this.termBucketChange.emit(this.formBucketTerms.value);
      this.errorsCount.emit(this.countErrors());
    });
  }

  initFormBucketTerms() {
    this.formBucketTerms = this.fb.group({
      sortBy: [this.fields[0].field],
      asc: [false],
      size: [5, Validators.required],
    });
  }

  loadFieldSort(metrics: MetricAggregationType[]) {
    this.fields = [];
    for (const metric of metrics) {
      this.fields.push(
        {
          field: (metric.aggregation === 'COUNT' ? '_count' : metric.id).toString(),
          fieldName: 'Metric: ' + metric.aggregation + ' (' + (metric.aggregation === 'COUNT' ? '_count' : metric.field) + ')'
        });
    }
    this.fields.push({field: '_key', fieldName: 'Alphabetical'});
  }

  countErrors(): number {
    return getErrorCountForm(this.formBucketTerms);
  }
}
