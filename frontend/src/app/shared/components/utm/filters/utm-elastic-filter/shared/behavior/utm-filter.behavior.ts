import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';
import {ElasticFilterType} from '../../../../../../types/filter/elastic-filter.type';

@Injectable({providedIn: 'root'})
export class UtmFilterBehavior {
  /**
   * Use this to add/remove filter
   */
  $filterChange = new BehaviorSubject<ElasticFilterType>(null);
  /**
   * Use to add exist filter on add field to table
   */
  $filterExistChange = new BehaviorSubject<ElasticFilterType>(null);

  processKey(key: string): string {
    const regex = /\.(\d+)\./g;
    if (regex.test(key)) {
      return key.replace(regex, '.');
    } else if (!isNaN(Number(key.substring(key.lastIndexOf('.'), key.length)))) {
      return key.substring(0, key.lastIndexOf('.'));
    } else {
      return key;
    }
  }
}
