import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormArray, FormBuilder, FormGroup, Validators} from '@angular/forms';
import {Observable} from 'rxjs';
import {BucketTermsType} from '../../../../shared/chart/types/metric/bucket-terms.type';
import {MetricBucketsType} from '../../../../shared/chart/types/metric/metric-buckets.type';
import {VisualizationType} from '../../../../shared/chart/types/visualization.type';
import {ChartTypeEnum} from '../../../../shared/enums/chart-type.enum';
import {FieldDataService} from '../../../../shared/services/elasticsearch/field-data.service';
import {ElasticSearchFieldInfoType} from '../../../../shared/types/elasticsearch/elastic-search-field-info.type';
import {getErrorCountFormArray} from '../shared/functions/form-error-count';
import {filterFieldAgg} from '../shared/functions/util-field';

@Component({
  selector: 'app-list-columns',
  templateUrl: './list-columns.component.html',
  styleUrls: ['./list-columns.component.scss']
})
export class ListColumnsComponent implements OnInit {
  @Input() mode: string;
  @Input() chart: ChartTypeEnum;
  @Input() visualization: VisualizationType;
  @Output() bucketChange = new EventEmitter<MetricBucketsType>();
  @Output() bucketErrors = new EventEmitter<boolean>();
  aggregation: string;
  listForm: FormGroup;
  fields: ElasticSearchFieldInfoType[];
  allFields: ElasticSearchFieldInfoType[];
  // count error for bucket aggregation
  errorTerms = 0;
  viewBucket = -1;
  chartTypeEnum = ChartTypeEnum;
  bucketId = 1000;

  constructor(private fieldDataBehavior: FieldDataService,
              private fb: FormBuilder) {
  }

  get columns() {
    return this.listForm.controls.columns as FormArray;
  }

  ngOnInit() {
    this.initFormColumns();
    this.fieldDataBehavior.$fields.subscribe(data => {
      this.allFields = filterFieldAgg(data);
      this.fields = filterFieldAgg(data);
    });
    if (this.mode === 'edit') {
      if (this.visualization.aggregationType.bucket) {
        this.convertObjectToArray(this.visualization.aggregationType.bucket).subscribe(() => {
          this.bucketChange.emit(this.buildBucketObject());
        });
      }
    }
    this.listForm.valueChanges.subscribe(() => {
      this.bucketChange.emit(this.buildBucketObject());
      this.bucketErrors.emit(this.countErrors() === 0);
    });
  }

  initFormColumns() {
    this.listForm = this.fb.group({
      columns: this.fb.array([]),
    });
  }

  addColumn(bucketType: string) {
    this.columns.push(this.fb.group({
      id: [this.bucketId, Validators.required],
      enable: [true],
      aggregation: ['TERMS', Validators.required],
      field: ['', Validators.required],
      customLabel: '',
      type: [bucketType],
      subBucket: [],
      terms: [new BucketTermsType()],
      significantTerms: [],
      dateHistogram: []
    }));
    this.viewBucket = this.columns.length - 1;
    this.bucketId += 1;
  }

  addColumnEdit(bucket: MetricBucketsType) {
    this.columns.push(this.fb.group({
      id: [bucket.id, Validators.required],
      enable: [bucket.enable],
      aggregation: [bucket.aggregation, Validators.required],
      field: [bucket.field, Validators.required],
      customLabel: bucket.customLabel,
      type: [bucket.type],
      subBucket: [bucket.subBucket],
      terms: [bucket.terms],
      significantTerms: [bucket.significantTerms],
      dateHistogram: [bucket.dateHistogram]
    }));
  }

  convertObjectToArray(bucket: MetricBucketsType): Observable<string> {
    return new Observable<string>(subscriber => {
      while (bucket) {
        if (typeof bucket.id === 'string') {
          bucket.id = Number(bucket.id);
        }
        this.bucketId = bucket.id + 1;
        this.addColumnEdit(bucket);
        bucket = bucket.subBucket;
      }
      subscriber.next('finish');
    });
  }

  deleteColumn(index) {
    this.columns.removeAt(index);
    if (!isNaN(index - 1)) {
      this.columns.at(index - 1).get('subBucket').setValue(null);
    }
  }

  viewColumnDetail(index) {
    if (this.viewBucket !== index) {
      this.viewBucket = index;
    } else {
      this.viewBucket = -1;
    }
  }

  countErrors(): number {
    return getErrorCountFormArray(this.columns) + this.errorTerms;
  }

  /**
   * Build object bucket hierarchical
   */
  private buildBucketObject(): MetricBucketsType {
    const arr: MetricBucketsType[] = this.columns.value ? this.columns.value : [];
    if (arr.length > 1) {
      for (let i = 0; i < arr.length; i++) {
        if (arr[i + 1] !== undefined) {
          arr[i].subBucket = arr[i + 1];
        }
      }
    }
    return arr[0];
  }


}
