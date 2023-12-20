import {Injectable} from '@angular/core';
import {LocalStorageService} from 'ngx-webstorage';
import {ElasticSearchFieldInfoType} from '../../types/elasticsearch/elastic-search-field-info.type';

@Injectable({
  providedIn: 'root'
})
/**
 * Use this clase to retrive field for index pattern from local storage, to improve speed load
 */
export class LocalFieldService {

  constructor(private localStorage: LocalStorageService) {
  }

  getPatternStoredFields(indexPattern: string): ElasticSearchFieldInfoType[] {
    return this.localStorage.retrieve(indexPattern + INDEX_PATTERN_FIELD);
  }

  setPatternStoredFields(indexPattern: string, fields: ElasticSearchFieldInfoType[]) {
    this.localStorage.store(indexPattern + INDEX_PATTERN_FIELD, fields);
  }

}

export const INDEX_PATTERN_FIELD = '_fields';
