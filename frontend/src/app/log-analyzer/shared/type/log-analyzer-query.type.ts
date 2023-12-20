import {DataNatureTypeEnum} from '../../../shared/enums/nature-data.enum';
import {ElasticFilterType} from '../../../shared/types/filter/elastic-filter.type';
import {UtmIndexPattern} from '../../../shared/types/index-pattern/utm-index-pattern';
import {UtmFieldType} from '../../../shared/types/table/utm-field.type';

export class LogAnalyzerQueryType {
  columnsType?: UtmFieldType[];
  filtersType?: ElasticFilterType[];
  creationDate?: Date;
  description?: string;
  dataOrigin?: DataNatureTypeEnum;
  id?: number;
  modificationDate?: Date;
  name?: string;
  owner?: string;
  pattern?: UtmIndexPattern;
  idPattern?: number;
}
