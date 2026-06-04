import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {BehaviorSubject, interval, Observable, of, Subject, Subscription} from 'rxjs';
import {catchError, distinctUntilChanged, filter, first, map, takeUntil, tap} from 'rxjs/operators';
import {SERVER_API_URL} from '../../app.constants';

@Injectable({
  providedIn: 'root'
})
export class ApiServiceCheckerService {

  public resourceUrl = SERVER_API_URL + 'api/ping';
  private retryInterval = 5000;

  private isOnline = new BehaviorSubject<boolean>(false);
  public isOnlineApi$ = this.isOnline.asObservable();

  private stopInterval$ = new Subject<void>();
  private intervalSub?: Subscription;

  constructor(private http: HttpClient) { }

  init() {
    this.checkApiAvailability();

    this.isOnlineApi$
      .pipe(distinctUntilChanged())
      .subscribe(isOnline => {
        if (!isOnline) {
          this.startCheckApiIsOnline();
        } else {
          this.stopChecking();
        }
      });
  }

  private checkApiAvailability() {
    this.checkApiStatus()
      .subscribe(status => this.isOnline.next(status));
  }

  private checkApiStatus(): Observable<boolean> {
    return this.http.get(this.resourceUrl, { observe: 'response' }).pipe(
      map(res => res.status === 200),
      catchError(() => of(false))
    );
  }

  private startCheckApiIsOnline() {
    if (this.intervalSub) { return; }

    this.intervalSub = interval(this.retryInterval)
      .pipe(takeUntil(this.stopInterval$))
      .subscribe(() => {
        this.checkApiStatus().
        subscribe(status => this.isOnline.next(status));
      });
  }

  private stopChecking() {
    this.stopInterval$.next();
    this.intervalSub.unsubscribe();
    this.intervalSub = undefined;
  }
}
