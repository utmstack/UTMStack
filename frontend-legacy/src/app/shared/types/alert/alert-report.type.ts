import {ElasticFilterType} from '../filter/elastic-filter.type';
import {UtmFieldType} from '../table/utm-field.type';

export interface IAlertReportType {
  id?: number;
  repDate?: Date;
  repDescription?: string;
  filters?: ElasticFilterType[];
  columns?: UtmFieldType[];
  repLimit?: number;
  repName?: string;
  userId?: number;
}

export class AlertReportType implements IAlertReportType {
  constructor(public  id?: number,
              public repDate?: Date,
              public repDescription?: string,
              public columns?: UtmFieldType[],
              public filters?: ElasticFilterType[],
              public repLimit?: number,
              public repName?: string,
              public userId?: number) {
    this.id = id ? id : null;
    this.repDate = repDate ? repDate : null;
    this.repDescription = repDescription ? repDescription : null;
    this.filters = filters ? filters : null;
    this.columns = columns ? columns : null;
    this.repLimit = repLimit ? repLimit : null;
    this.repName = repName ? repName : null;
    this.userId = userId ? userId : null;
  }
}
