
import {HttpResponse} from '@angular/common/http';
import { Injectable } from '@angular/core';
import {ActivatedRouteSnapshot, Resolve, RouterStateSnapshot} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {Observable, of} from 'rxjs';
import {catchError, map, tap} from 'rxjs/operators';
import {UtmServerService} from '../shared/services/utm-server.service';
import {UtmServerType} from '../shared/type/utm-server.type';
import {UtmToastService} from '../../shared/alert/utm-toast.service';

@Injectable()
export class ModuleResolverService implements Resolve<Observable<UtmServerType>> {

  constructor(private utmServerService: UtmServerService,
              private utmToastService: UtmToastService,
              private spinner: NgxSpinnerService) {}
  resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<UtmServerType> {
    this.spinner.show('loadingSpinner');
    return this.utmServerService.query({page: 0, size: 100})
      .pipe(
        tap(() => this.spinner.hide('loadingSpinner')),
        catchError(error => {
          console.log(error);
          this.utmToastService.showError('Failed to fetch servers',
            'An error occurred while fetching module data.');
          return of(null);
        }),
        map((resp: HttpResponse<UtmServerType[]>) => resp.body[0]));
  }
}
