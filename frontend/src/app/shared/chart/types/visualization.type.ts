import {ChartTypeEnum} from '../../enums/chart-type.enum';
import {DataNatureTypeEnum} from '../../enums/nature-data.enum';
import {ElasticFilterType} from '../../types/filter/elastic-filter.type';
import {UtmIndexPattern} from '../../types/index-pattern/utm-index-pattern';
import {MetricDataType} from './metric/metric-data.type';
import {ChartBuilderQueryLanguageEnum} from "../../enums/chart-builder-query-language.enum";

export class VisualizationType {
  chartConfig?: any;
  chartType?: ChartTypeEnum;
  description?: string;
  filterType?: ElasticFilterType[];
  id?: number;
  idPattern?: number;
  modifiedDate?: string;
  eventType?: DataNatureTypeEnum;
  name?: string;
  pattern?: UtmIndexPattern;
  queryType?: any;
  userCreated?: number;
  userModified?: number;
  aggregationType?: MetricDataType;
  chartAction?: any;
  showTime?: boolean;
  systemOwner?: boolean;
  page?: any;
  sqlQuery?: string;
  queryLanguage?: ChartBuilderQueryLanguageEnum;
}

