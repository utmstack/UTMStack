import { Injectable } from '@angular/core';
import { BehaviorSubject, interval, Observable, Subscription } from 'rxjs';
import { filter } from 'rxjs/operators';

export enum RefreshType {
  ALL = 'ALL',
  CHART_ALERT_BY_STATUS = 'CHART_ALERT_BY_STATUS',
  CHART_COMMON_PIE_SEVERITY = 'CHART_COMMON_PIE_SEVERITY',
  CHART_ALERT_BY_CATEGORY = 'CHART_ALERT_BY_CATEGORY',
  CHART_COMMON_TABLE_TOP_ALERT = 'CHART_COMMON_TABLE_TOP_ALERT',
  CHART_COMMON_PIE_EVENT = 'CHART_COMMON_PIE_EVENT',
  CHART_COMMON_TABLE_TOP_EVENT = 'CHART_COMMON_TABLE_TOP_EVENT',
  CHART_EVENT_IN_TIME = 'CHART_EVENT_IN_TIME',
}

@Injectable({
  providedIn: 'root'
})
export class RefreshService {
  private refreshSubject = new BehaviorSubject<string>(null);
  private refreshInterval = 10000;
  private interval$: Observable<number>;
  private subscription: Subscription;

  constructor() {
    this.interval$ = interval(this.refreshInterval);
  }

  startInterval() {
    this.subscription = this.interval$
      .subscribe(() => {
        console.log('emitting refresh');
        this.sendRefresh();
      });
  }

  get refresh$() {
    return this.refreshSubject.asObservable()
      .pipe(filter(value => !!value));
  }

  sendRefresh(type: RefreshType = RefreshType.ALL){
    this.refreshSubject.next(type);
  }

  stopInterval() {
    if (this.subscription) {
      this.subscription.unsubscribe();
      this.refreshSubject.next(null);
      console.log('refresh stopped');
    }
  }

  setRefreshInterval(refreshInterval: number) {
    this.refreshInterval = refreshInterval;
    this.interval$ = interval(this.refreshInterval);
    this.startInterval();
  }
}
