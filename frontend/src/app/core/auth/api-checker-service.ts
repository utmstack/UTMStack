import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {BehaviorSubject, interval, Observable, Subject} from 'rxjs';
import {first, takeUntil} from 'rxjs/operators';
import {SERVER_API_URL} from '../../app.constants';

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

  setOnlineStatus(status: boolean) {
    this.isOnline.next(status);
  }
}
