import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormArray, FormBuilder, FormGroup, Validators} from '@angular/forms';
import {Observable} from 'rxjs';
import {BucketTermsType} from '../../../../shared/chart/types/metric/bucket-terms.type';
import {MetricBucketsType} from '../../../../shared/chart/types/metric/metric-buckets.type';
import {VisualizationType} from '../../../../shared/chart/types/visualization.type';
import {MULTIPLE_BUCKETS_CHART} from '../../../../shared/constants/visualization-bucket-metric.constant';
import {ChartTypeEnum} from '../../../../shared/enums/chart-type.enum';
import {ElasticDataTypesEnum} from '../../../../shared/enums/elastic-data-types.enum';
import {FieldDataService} from '../../../../shared/services/elasticsearch/field-data.service';
import {ElasticSearchFieldInfoType} from '../../../../shared/types/elasticsearch/elastic-search-field-info.type';
import {BUCKETS_AGGREGATIONS} from '../../../shared/const/aggregation.const';
import {BucketAggregationEnum} from '../../../shared/enums/bucket-aggregation.enum';
import {filterFieldDataType} from '../../../shared/util/elasticsearch/elastic-field.util';
import {getErrorCountFormArray} from '../shared/functions/form-error-count';
import {filterFieldAgg} from '../shared/functions/util-field';

@Component({
  selector: 'app-bucket-aggregation',
  templateUrl: './bucket-aggregation.component.html',
  styleUrls: ['./bucket-aggregation.component.scss']
})
export class BucketAggregationComponent implements OnInit {
  @Input() mode: string;
  @Input() chart: ChartTypeEnum;
  @Input() visualization: VisualizationType;
  @Output() bucketChange = new EventEmitter<MetricBucketsType>();
  @Output() bucketErrors = new EventEmitter<boolean>();
  aggregations: { agg: string, value: string }[] = BUCKETS_AGGREGATIONS;
  aggregation: string;
  bucketForm: FormGroup;
  bucketAggregationEnum = BucketAggregationEnum;
  fields: ElasticSearchFieldInfoType[];
  allFields: ElasticSearchFieldInfoType[];
  // count error for bucket aggregation
  errorTerms = 0;
  errorSigTerms = 0;
  errorDateHistogram = 0;
  bucketAdded = false;
  viewBucket = -1;
  chartTypeEnum = ChartTypeEnum;
  multipleBuckets = MULTIPLE_BUCKETS_CHART;
  bucketId = 1000;

  constructor(private fieldDataBehavior: FieldDataService,
              private fb: FormBuilder) {
  }

  get buckets() {
    return this.bucketForm.controls.buckets as FormArray;
  }

  ngOnInit() {
    // Validate chart type for apply histogram agg
    if (!this.multipleBuckets.includes(this.chart)) {
      this.aggregations.filter(value => value.agg !== BucketAggregationEnum.DATE_HISTOGRAM);
    }
    this.initFormBucket();
    this.fieldDataBehavior.$fields.subscribe(data => {
      this.allFields = filterFieldAgg(data);
      this.fields = filterFieldAgg(data);
    });
    if (this.mode === 'edit') {
      if (this.visualization.aggregationType.bucket) {
        this.convertObjectToArray(this.visualization.aggregationType.bucket).subscribe(() => {
          this.bucketAdded = true;
          this.bucketChange.emit(this.buildBucketObject());
        });
      }
    }
    this.bucketForm.valueChanges.subscribe(() => {
      this.bucketChange.emit(this.buildBucketObject());
      this.bucketErrors.emit(this.countErrors() === 0);
    });
  }

  initFormBucket() {
    this.bucketForm = this.fb.group({
      buckets: this.fb.array([]),
    });
  }

  addBucket(bucketType: string) {
    this.buckets.push(this.fb.group({
      id: [this.bucketId, Validators.required],
      enable: [true],
      aggregation: ['', Validators.required],
      field: ['', Validators.required],
      customLabel: '',
      type: [bucketType],
      subBucket: [],
      terms: [new BucketTermsType()],
      significantTerms: [],
      dateHistogram: []
    }));
    this.viewBucket = this.buckets.length - 1;
    this.bucketId += 1;
  }

  addBucketEdit(bucket: MetricBucketsType) {
    this.buckets.push(this.fb.group({
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
        this.addBucketEdit(bucket);
        bucket = bucket.subBucket;
      }
      subscriber.next('finish');
    });
  }

  deleteBucket(index) {
    this.buckets.removeAt(index);
    if (!isNaN(index - 1)) {
      this.buckets.at(index - 1).get('subBucket').setValue(null);
    }
    if (this.buckets.length === 0) {
      this.bucketAdded = false;
    }
  }

  viewBucketDetail(index) {
    if (this.viewBucket !== index) {
      this.viewBucket = index;
    } else {
      this.viewBucket = -1;
    }
  }

  changeFieldType($event, index) {
    switch ($event.value) {
      case this.bucketAggregationEnum.DATE_HISTOGRAM:
        this.buckets.at(index).get('terms').setValue(null);
        this.extractFields(ElasticDataTypesEnum.DATE);
        break;
      case this.bucketAggregationEnum.TERMS:
        this.buckets.at(index).get('dateHistogram').setValue(null);
        this.fields = this.allFields;
        break;
      case this.bucketAggregationEnum.SIGNIFICANT_TERMS:
        this.extractFields(ElasticDataTypesEnum.TEXT);
        break;
    }
  }

  extractFields(type: string) {
    filterFieldDataType(this.fields, type).subscribe(f => {
      this.fields = f;
    });
  }

  countErrors(): number {
    return getErrorCountFormArray(this.buckets) + this.errorTerms + this.errorSigTerms + this.errorDateHistogram;
  }

  errorCountAggTermsChange($event: number) {
    this.errorTerms = $event;
  }

  errorCountAggSigTermsChange($event: number) {
    this.errorSigTerms = $event;
  }

  errorCountDateHistogramChange($event: number) {
    this.errorDateHistogram = $event;
  }

  viewSplice(bucketType: string) {
    this.bucketAdded = true;
    if (this.buckets.length === 0) {
      this.addBucket(bucketType);
    }
  }

  // changeAllFieldTextToKeyword(): ElasticSearchFieldInfoType[] {
  //   const fieldKeyword: ElasticSearchFieldInfoType[] = [];
  //   for (const field of this.allFields) {
  //     if (field.type === 'text' && !field.name.includes('.keyword')) {
  //       field.name = field.name + '.keyword';
  //     }
  //     fieldKeyword.push(field);
  //   }
  //   return fieldKeyword;
  // }

  changeFieldTextToKeyword($event, index: number) {
    if ($event.type === ElasticDataTypesEnum.TEXT) {
      this.buckets.at(index).get('field').setValue($event.name + '.keyword');
    }
  }

  shouldApplyAxis() {
    return ((this.visualization.chartType === this.chartTypeEnum.AREA_LINE_CHART ||
      this.visualization.chartType === this.chartTypeEnum.LINE_CHART ||
      this.visualization.chartType === this.chartTypeEnum.BAR_CHART ||
      this.visualization.chartType === this.chartTypeEnum.BAR_HORIZONTAL_CHART ||
      this.visualization.chartType === this.chartTypeEnum.HEATMAP_CHART) &&
      this.buckets.controls.findIndex(value =>
        value.get('type').value === 'AXIS') === -1) || this.buckets.length === 0;
  }

  move(shift, currentIndex) {
    let newIndex: number = currentIndex + shift;
    if (newIndex === -1) {
      newIndex = this.buckets.length - 1;
    } else if (newIndex === this.buckets.length) {
      newIndex = 0;
    }
    const currentGroup = this.buckets.at(currentIndex);
    this.buckets.removeAt(currentIndex);
    this.buckets.insert(newIndex, currentGroup);
  }

  /**
   * Build object bucket hierarchical
   */
  private buildBucketObject(): MetricBucketsType {
    const arr: MetricBucketsType[] = this.buckets.value ? this.buckets.value : [];
    if (arr.length > 1) {
      for (let i = 0; i < arr.length; i++) {
        if (arr[i + 1] !== undefined) {
          arr[i].subBucket = arr[i + 1];
        }
      }
    }
    return arr[0];
  }

  onTermBucketChange($event: BucketTermsType, index: number) {
    this.buckets.at(index).get('terms').setValue($event);
  }

}
