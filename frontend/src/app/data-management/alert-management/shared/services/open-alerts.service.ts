import { Injectable, OnDestroy } from '@angular/core';
import {LocalStorageService} from 'ngx-webstorage';
import { BehaviorSubject, interval, Subject } from 'rxjs';
import { catchError, switchMap, takeUntil, tap } from 'rxjs/operators';
import { AlertOpenStatusService } from '../../../../shared/webflux/alert-open-status.service';

export const OPEN_ALERTS_KEY = 'open-alerts';

@Injectable({
  providedIn: 'root',
})
export class OpenAlertsService implements OnDestroy {
  private openAlertsBehaviorSubject = new BehaviorSubject<number | null>(null);
  openAlerts$ = this.openAlertsBehaviorSubject.asObservable();
  private destroy$ = new Subject<void>();
  _openAlerts = 0;

  constructor(private alertOpenStatusService: AlertOpenStatusService,
              private localStorage: LocalStorageService) {

    const openAlerts = this.localStorage.retrieve(OPEN_ALERTS_KEY);
    if (!!openAlerts) {
      this._openAlerts = openAlerts;
      this.openAlertsBehaviorSubject.next(openAlerts);
    }
    this.startInterval();
  }

  private startInterval() {
    interval(3000)
      .pipe(
        takeUntil(this.destroy$),
        switchMap(() => this.alertOpenStatusService.getOpenAlert()),
        tap((response) => {
          this.openAlerts = response.body;
          this.localStorage.store(OPEN_ALERTS_KEY, this.openAlerts);
          if (this.openAlerts > 0 && this.openAlerts !== this.openAlerts) {
            this.openAlertsBehaviorSubject.next(this.openAlerts);
          }
        }),
        catchError((err) => {
          console.error('Error fetching alerts:', err);
          return [];
        })
      )
      .subscribe();
  }

  stopInterval() {
    this.destroy$.next();
    this.destroy$.complete();
  }

  reset() {
    this.openAlertsBehaviorSubject.next(0);
  }

  get openAlerts(){
    return this._openAlerts;
  }

  set openAlerts(openAlerts) {
    this._openAlerts = openAlerts;
  }

  ngOnDestroy() {
    this.stopInterval();
  }
}
