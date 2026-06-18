import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {BehaviorSubject, Observable, of} from 'rxjs';
import {catchError, map, shareReplay, tap} from 'rxjs/operators';
import {SERVER_API_URL} from '../../app.constants';
import {FederationMode} from '../domain/federation-mode.model';

@Injectable({providedIn: 'root'})
export class FederationModeService {
  private readonly endpoint = SERVER_API_URL + 'api/v1/mode';
  private modeSubject = new BehaviorSubject<FederationMode | null>(null);
  private detect$: Observable<FederationMode> | null = null;

  readonly mode$ = this.modeSubject.asObservable();
  readonly active$ = this.modeSubject.pipe(map(m => !!m && m.federation === true));

  constructor(private http: HttpClient) {}

  detect(): Observable<FederationMode> {
    if (!this.detect$) {
      this.detect$ = this.http.get<FederationMode>(this.endpoint).pipe(
        catchError(() => of<FederationMode>({federation: false, version: ''})),
        tap(mode => this.modeSubject.next(mode)),
        shareReplay(1)
      );
    }
    return this.detect$;
  }

  get current(): FederationMode | null {
    return this.modeSubject.getValue();
  }

  get isActive(): boolean {
    const current = this.modeSubject.getValue();
    return !!current && current.federation === true;
  }
}
