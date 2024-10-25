import {Injectable} from '@angular/core';
import {BehaviorSubject, interval, Subscription} from 'rxjs';
import {map, switchMap, tap} from 'rxjs/operators';
import {AlertOpenStatusService} from '../../../../shared/webflux/alert-open-status.service';

@Injectable({
  providedIn: 'root'
})
export class OpenAlertsService {
  private openAlertsBehaviorSubject = new BehaviorSubject<number>(0);
  openAlerts$ = this.openAlertsBehaviorSubject.asObservable();
  private interval$ = interval(30000);
  private subscription: Subscription;
  private readonly timeOutId: number;

  constructor(private alertOpenStatusService: AlertOpenStatusService) {
    this.timeOutId = setTimeout(() => {
      this.fetchOpenAlerts();
    }, 30000);
  }

  fetchOpenAlerts() {
    this.subscription = this.interval$.pipe(
      switchMap( () => this.alertOpenStatusService.getOpenAlert()),
      map((response) => response.body),
      tap((openAlerts) => {
        if (openAlerts > 0 && openAlerts !== this.openAlertsBehaviorSubject.value) {
          this.openAlertsBehaviorSubject.next(openAlerts);
        }
      })
    ).subscribe();
  }

  stopInterval(){
    if (this.subscription) {
      this.subscription.unsubscribe();
    }

    if (this.timeOutId) {
      clearTimeout(this.timeOutId);
    }
  }

  reset(){
    this.openAlertsBehaviorSubject.next(0);
  }
}
