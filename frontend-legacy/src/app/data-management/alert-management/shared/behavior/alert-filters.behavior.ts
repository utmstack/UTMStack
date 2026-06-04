import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';
import {ElasticFilterType} from '../../../../shared/types/filter/elastic-filter.type';

@Injectable({
  providedIn: 'root'
})
export class AlertFiltersBehavior {
  /**
   * Use tp have control on current filters
   */
  $filters = new BehaviorSubject<ElasticFilterType[]>(null);
  /**
   * Use to delete dynamic filters
   */
  $deleteFilterValue = new BehaviorSubject<{ field: string, value: any }>(null);

  $resetFilter = new BehaviorSubject<boolean>(null);
}
