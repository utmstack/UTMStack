import {DataType} from '../../../rule-management/models/rule.model';

export interface LogstashFilterType {
  id?: number;
  filterName?: string;
  logstashFilter?: string;
  systemOwner?: boolean;
  filterVersion?: string;
  datatype: DataType;
}
