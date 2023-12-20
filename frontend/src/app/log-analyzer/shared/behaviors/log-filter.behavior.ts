import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';
import {ElasticFilterType} from '../../../shared/types/filter/elastic-filter.type';

@Injectable({providedIn: 'root'})
export class LogFilterBehavior {
  $logFilter = new BehaviorSubject<{ filter: ElasticFilterType[], sort: string }>(null);
}



