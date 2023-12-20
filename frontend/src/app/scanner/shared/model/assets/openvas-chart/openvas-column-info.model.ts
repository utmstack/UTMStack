import {OpenvasAggregateColumnModel} from './openvas-aggregate-column.model';

export interface IOpenvasColumnInfoModel {
  aggregateColumn?: OpenvasAggregateColumnModel;
}

export class OpenvasColumnInfoModel implements IOpenvasColumnInfoModel {
  constructor(aggregateColumn?: OpenvasAggregateColumnModel) {
  }
}
