import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';
import {ElasticFilterType} from '../types/filter/elastic-filter.type';
import {Menu} from '../types/menu/menu.model';

@Injectable({providedIn: 'root'})
export class DashboardBehavior {
  $dashboard = new BehaviorSubject<Menu>(null);
  $filterDashboard = new BehaviorSubject<{
    indexPattern: string, filter: ElasticFilterType[]
  }>(null);
}
