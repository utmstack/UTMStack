import {AlertTags} from '../../../shared/types/alert/alert-tag.type';
import {ElasticFilterType} from '../../../shared/types/filter/elastic-filter.type';

export class AlertRuleType {
  id: number;
  name: string;
  description: string;
  conditions: ElasticFilterType[];
  tags: AlertTags[];
  createdBy: string;
  createdDate: Date;
  lastModifiedBy: string;
  deleted: false;
  lastModifiedDate: Date;
}
