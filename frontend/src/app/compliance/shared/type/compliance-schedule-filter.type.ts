import {ElasticFilterType} from '../../../shared/types/filter/elastic-filter.type';

export class ComplianceScheduleFilterType {
  indexPattern: string;
  filterType: ElasticFilterType[];
}
