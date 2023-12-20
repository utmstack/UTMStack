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
}
