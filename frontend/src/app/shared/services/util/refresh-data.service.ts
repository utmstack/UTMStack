import { BehaviorSubject, Observable } from 'rxjs';

export abstract class RefreshDataService<T, X> {
  private refreshSubject = new BehaviorSubject<T>(null);

  get onRefresh$(): Observable<T> {
    return this.refreshSubject.asObservable();
  }
  notifyRefresh(value: T): void {
    this.refreshSubject.next(value);
  }

  abstract fetchData(request: any): Observable<X>;
}
