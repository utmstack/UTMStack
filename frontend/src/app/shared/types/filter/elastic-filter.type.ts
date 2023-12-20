import {UtmIndexPattern} from '../index-pattern/utm-index-pattern';

export class ElasticFilterType {
  label?: string;
  pattern?: string;
  field?: string;
  value?: any;
  operator?: any;
  status?: 'ACTIVE' | 'INACTIVE';
}
