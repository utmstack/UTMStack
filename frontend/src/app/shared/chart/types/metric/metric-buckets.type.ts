import {BucketDateHistogramType} from './bucket-date-histogram.type';
import {BucketSignificantTermsType} from './bucket-significant-terms.type';
import {BucketTermsType} from './bucket-terms.type';

export class MetricBucketsType {
  id: number;
  enable?: boolean;
  aggregation: string;
  field: string;
  customLabel?: string;
  subBucket?: MetricBucketsType;
  terms?: BucketTermsType;
  significantTerms?: BucketSignificantTermsType;
  dateHistogram?: BucketDateHistogramType;
  type: 'axis' | 'bucket';
}
