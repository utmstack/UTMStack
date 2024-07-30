import {Injectable} from '@angular/core';
import {BehaviorSubject, interval, Subject, Subscription} from 'rxjs';
import {LocalStorageService} from 'ngx-webstorage';
import {filter, takeUntil} from "rxjs/operators";

@Injectable({
  providedIn: 'root'
})
export class RefreshService {
  private refreshSubject = new BehaviorSubject<string>(null);
  private refreshInterval = 10000; // 10 seconds
  private destroy$: Subject<void> = new Subject<void>();

  constructor(private localStorageService: LocalStorageService) {
  }

  startInterval() {
    interval(this.refreshInterval)
      .pipe(takeUntil(this.destroy$))
      .subscribe(() => {
        console.log('emitting refresh');
        this.refreshSubject.next('refresh');
    });
  }

  get refresh$() {
    return this.refreshSubject.asObservable()
      .pipe(filter(value => !!value));
  }

  stopRefresh() {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
