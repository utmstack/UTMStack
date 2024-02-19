import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {BehaviorSubject, interval, Observable, of, Subject, throwError} from 'rxjs';
import {catchError, delay, first, switchMap, takeUntil, takeWhile, tap} from 'rxjs/operators';
import {SERVER_API_URL} from "../../app.constants";

@Injectable({
  providedIn: 'root'
})
export class ApiServiceCheckerService {

  public resourceUrl = SERVER_API_URL + 'api/ping';
  private retryInterval = 5000;
  private isOnline: BehaviorSubject<boolean> = new BehaviorSubject<boolean>(null);
  public isOnlineApi$: Observable<boolean> = this.isOnline.asObservable();
  private stopInterval$: Subject<boolean> = new Subject<boolean>();

  constructor(private http: HttpClient) {
  }

  checkApiAvailability() {
    interval(this.retryInterval)
      .pipe(takeUntil(this.stopInterval$))
      .subscribe(() => {
      this.http
        .get(this.resourceUrl, { observe: 'response' })
        .pipe(first())
        .subscribe(
          resp => {
            if (resp.status === 200) {
              this.isOnline.next(true);
              this.stopInterval$.next(true);
            } else {
              this.isOnline.next(false);
            }
          },
          err => this.isOnline.next(false)
        );
    });
  }
}

/*return this.http.get(this.resourceUrl).pipe(
  catchError(() => {
    return of(false);
  }),
  switchMap((result: any) => {
    if (result != null) {
      return of(true);
    }
  }),
  takeWhile((isAvailable: any) => isAvailable != null)
);
}*/
