import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';
import {ElasticFilterType} from '../../../../shared/types/filter/elastic-filter.type';

@Injectable({
  providedIn: 'root'
})
export class FileFiltersBehavior {
  /**
   * Use tp have control on current filters
   */
  $fileFilters = new BehaviorSubject<ElasticFilterType[]>(null);
  /**
   * Use to delete dynamic filters
   */
  $fileDeleteFilterValue = new BehaviorSubject<{ field: string, value: any }>(null);

  $fileResetFilter = new BehaviorSubject<boolean>(null);
}
