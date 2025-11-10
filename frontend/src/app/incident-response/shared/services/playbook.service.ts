import { Injectable } from '@angular/core';
import { BehaviorSubject, of, Subject } from 'rxjs';
import { catchError, filter, finalize, map, switchMap } from 'rxjs/operators';
import { UtmToastService } from '../../../shared/alert/utm-toast.service';
import { IncidentResponseRuleService } from './incident-response-rule.service';

@Injectable()
export class PlaybookService {
  private request$ = new Subject<any>();
  private loading = new BehaviorSubject<boolean>(false);
  loading$ = this.loading.asObservable();
  private totalItems = new BehaviorSubject<number>(null);
  totalItems$ = this.totalItems.asObservable();

  constructor(
    private incidentResponseRuleService: IncidentResponseRuleService,
    private utmToastService: UtmToastService
  ) {}

  playbooks$ = this.request$.pipe(
    filter(request =>  !!request),
    switchMap(request => {
      this.loading.next(true);
      return this.incidentResponseRuleService.query(request).pipe(
        map(response => {
          this.totalItems.next(Number(response.headers.get('X-Total-Count')));
          return response.body;
        }),
        catchError(error => {
          this.utmToastService.showError('Error', 'An error occurred while fetching playbooks.');
          return of([]);
        }),
        finalize(() => this.loading.next(false))
      );
    })
  );

  loadData(request: any){
    this.request$.next(request);
  }

  reset() {
    this.request$.next(null);
    this.totalItems.next(0);
  }
}
