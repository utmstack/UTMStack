import {
  ElasticFilterDefaultTime
} from '../../../shared/components/utm/filters/elastic-filter-time/elastic-filter-time.component';

export interface SkeletonInterface {
  data: any;
  uuid: string;
  defaultTime: ElasticFilterDefaultTime;
}
