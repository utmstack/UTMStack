import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';
import {MetricAggregationType} from '../../../../../shared/chart/types/metric/metric-aggregation.type';

@Injectable({
  providedIn: 'root'
})
export class MetricDataBehavior {
  $metric = new BehaviorSubject<MetricAggregationType[]>([]);
  $metricDeletedId = new BehaviorSubject<number>(-1);
}
