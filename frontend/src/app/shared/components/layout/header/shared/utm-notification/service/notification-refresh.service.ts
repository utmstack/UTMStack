import {Injectable} from '@angular/core';
import {BehaviorSubject, interval, Observable, Subscription} from 'rxjs';
import {filter} from 'rxjs/operators';

@Injectable()
export class NotificationRefreshService {
  private refreshNotificationBehavior = new BehaviorSubject<boolean>(null);
  private onScrollNotificationBehavior = new BehaviorSubject<boolean>(null);
  private subscription: Subscription;
  private interval$: Observable<number>;

  constructor() {
    this.setRefreshInterval(3000);
  }

  startInterval() {
    this.subscription = this.interval$
      .subscribe(() => {
        console.log('Notification refresh!!!');
        this.sendRefresh();
      });
  }
  get refresh$() {
    return this.refreshNotificationBehavior.asObservable()
      .pipe(filter(value => !!value));
  }

  get onScrollNotification$(){
    return this.onScrollNotificationBehavior.asObservable();
  }

  onScroll() {
    console.log('click');
    this.onScrollNotificationBehavior.next(true);
  }

  sendRefresh() {
    this.refreshNotificationBehavior.next(true);
  }

  stopInterval() {
    this.refreshNotificationBehavior.next(null);
    if (this.subscription) {
      this.subscription.unsubscribe();
      console.log('refresh stopped');
    }
  }

  setRefreshInterval(refreshInterval: number) {
    this.stopInterval();
    this.interval$ = interval(refreshInterval);
    this.startInterval();
  }
}
