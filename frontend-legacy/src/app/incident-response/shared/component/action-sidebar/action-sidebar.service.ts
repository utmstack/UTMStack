import {HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {BehaviorSubject, of} from 'rxjs';
import {catchError, filter, finalize, map, switchMap, tap} from 'rxjs/operators';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {
  IncidentResponseActionTemplate,
  IncidentResponseActionTemplateService
} from '../../services/incident-response-action-template.service';

@Injectable({
  providedIn: 'root'
})
export class ActionSidebarService {

  private request = new BehaviorSubject<any>(null);
  private loading = new BehaviorSubject<boolean>(false);

  request$ = this.request.asObservable();
  loading$ = this.loading.asObservable();

  actionTemplates$ = this.request$
    .pipe(
      filter(request => !!request),
      tap(() => this.loading.next(true)),
      switchMap((request) => this.incidentResponseActionTemplateService.query(request)
        .pipe(
          map((response: HttpResponse<IncidentResponseActionTemplate[]>) => response.body),
          catchError(() => {
            this.toastService.showError('Error', 'Failed to load action templates');
            return of([]);
          }),
          finalize(() => this.loading.next(false))
        )
      ),
    );

  constructor(private incidentResponseActionTemplateService: IncidentResponseActionTemplateService,
              private toastService: UtmToastService) {}

  loadData(request: any) {
    this.request.next(request);
  }

  reset() {
    this.request.next(null);
  }

}
