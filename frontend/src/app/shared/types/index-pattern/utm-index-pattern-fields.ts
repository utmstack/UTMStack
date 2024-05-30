import {ElasticSearchFieldInfoType} from '../elasticsearch/elastic-search-field-info.type';

export class UtmIndexPatternFields {
  indexPattern: string;
  fields: ElasticSearchFieldInfoType[];

  constructor(indexPattern: string, fields: ElasticSearchFieldInfoType[]) {
    this.indexPattern = indexPattern;
    this.fields = fields;
  }
}
