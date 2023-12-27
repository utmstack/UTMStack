import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {ElasticFilterType} from '../../../../shared/types/filter/elastic-filter.type';

@Injectable({
  providedIn: 'root'
})
export class AlertFilterManagementService {

  constructor() {
  }

  deleteFilterValue(field: string, value: any, filters: ElasticFilterType[]): Observable<ElasticFilterType[]> {
    return new Observable<ElasticFilterType[]>(subscriber => {
      const indexField = filters.findIndex(filter => filter.field === field);
      if (indexField !== -1) {
        const filterValueIndex = filters[indexField].value.findIndex(val => val = value);
        if (filterValueIndex !== -1) {
          if (typeof filters[indexField].value === 'string') {
            filters[indexField].value = [filters[indexField].value];
          }
          filters[indexField].value[filterValueIndex].splice(filterValueIndex, 1);
          subscriber.next(filters);
        }
      }
    });
  }


}
