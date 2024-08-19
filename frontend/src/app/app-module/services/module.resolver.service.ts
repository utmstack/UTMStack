import { Injectable } from '@angular/core';
import {ActivatedRouteSnapshot, Resolve, RouterStateSnapshot} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {Observable, of} from 'rxjs';
import {catchError, map, switchMap, tap} from 'rxjs/operators';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {UtmModulesEnum} from '../shared/enum/utm-module.enum';
import {UtmModulesService} from '../shared/services/utm-modules.service';
import {UtmServerService} from '../shared/services/utm-server.service';
import {UtmServerType} from '../shared/type/utm-server.type';

@Injectable()
export class ModuleResolverService implements Resolve<Observable<UtmServerType>> {
  utmModulesEnum = UtmModulesEnum;
  server: UtmServerType;

  constructor(private utmServerService: UtmServerService,
              private utmToastService: UtmToastService,
              private spinner: NgxSpinnerService,
              private utmModulesService: UtmModulesService) {}

  resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<any> {
    this.spinner.show('loadingSpinner');
    return this.utmServerService.query({page: 0, size: 100})
      .pipe(
        catchError(error => {
          console.log(error);
          this.utmToastService.showError('Failed to fetch servers',
            'An error occurred while fetching module data.');
          return of(null);
        }),
        tap((response) => this.server = response.body[0]),
        switchMap(response => this.getModules({
          page: 0,
          size: 100,
          'serverId.equals': response.body[0].id,
          sort: 'moduleCategory,asc'
        })));
  }

  getModules(req: any) {
    return this.utmModulesService.getModules(req)
      .pipe(
        map( response => {
          response.body.map(m => {
            if (m.moduleName === this.utmModulesEnum.BITDEFENDER) {
               m.prettyName = m.prettyName + ' GravityZone';
            }
          });
          return response.body;
        }),
        tap(() => this.spinner.hide('loadingSpinner')),
        catchError(error => {
          console.log(error);
          this.utmToastService.showError('Failed to fetch modules',
            'An error occurred while fetching module data.');
          return [];
        })
      );
  }
}
