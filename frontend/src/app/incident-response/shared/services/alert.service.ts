import {HttpErrorResponse, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {EMPTY, Observable} from 'rxjs';
import {catchError, map} from 'rxjs/operators';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {ElasticDataService} from '../../../shared/services/elasticsearch/elastic-data.service';
import {RefreshDataService} from '../../../shared/services/util/refresh-data.service';
import {sanitizeFilters} from '../../../shared/util/elastic-filter.util';

@Injectable(
  {
  providedIn: 'root'
  }
)
export class AlertService extends RefreshDataService<boolean, any[]> {

  constructor(private elasticDataService: ElasticDataService,
              private toastService: UtmToastService) {
    super();
  }

  fetchData(request: any): Observable<any[]> {
    return this.elasticDataService.search(
      request.page,
      request.size,
      100000000,
      request.dataNature,
      sanitizeFilters(request.filters),
      request.sort)
      .pipe(
        map((response: HttpResponse<any[]>) => response.body),
        catchError((err: HttpErrorResponse) => {
          this.toastService.showError('Error',
            'Unable to retrieve the list of alerts. Please try again or contact support.');
          return EMPTY;
        })
      );
  }
}
