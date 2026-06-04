import {Injectable} from '@angular/core';
import {LocalStorageService} from 'ngx-webstorage';
import {DataNatureTypeEnum} from '../../../shared/enums/nature-data.enum';
import {ElasticSearchFieldInfoType} from '../../../shared/types/elasticsearch/elastic-search-field-info.type';

@Injectable({
  providedIn: 'root'
})
export class LogAnalyzerFieldStoredService {

  prefix = '_fields';

  constructor(private localStorageService: LocalStorageService) {
  }

  setFieldStorage(dataNatureTypeEnum: DataNatureTypeEnum, fields: ElasticSearchFieldInfoType[]) {
    this.localStorageService.store((dataNatureTypeEnum + this.prefix), fields);
  }

  getFieldStorage(dataNatureTypeEnum: DataNatureTypeEnum): ElasticSearchFieldInfoType[] {
    return this.localStorageService.retrieve((dataNatureTypeEnum + this.prefix));
  }
}
