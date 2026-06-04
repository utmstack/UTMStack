import { Injectable } from '@angular/core';
import {BehaviorSubject, Observable, of} from 'rxjs';
import {catchError, filter, finalize, map, shareReplay, switchMap, tap} from 'rxjs/operators';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {UtmModulesEnum} from '../shared/enum/utm-module.enum';
import {UtmModulesService} from '../shared/services/utm-modules.service';
import {UtmServerService} from '../shared/services/utm-server.service';
import {UtmServerType} from '../shared/type/utm-server.type';

export const ModulesEnterprise = [
  UtmModulesEnum.MACOS,
  UtmModulesEnum.AS_400
];


export interface RequestModule {
  'moduleCategory.equals'?: string | null;
  'prettyName.contains'?: string | null;
  'serverId.equals'?: string | null;
  'moduleName.equals'?: string | null;
  sort?: string;
  page?: number;
  size?: number;
}

@Injectable()
export class ModuleService {
  utmModulesEnum = UtmModulesEnum;

  private requestBehaviorSubject = new BehaviorSubject<RequestModule>({});
  request$ = this.requestBehaviorSubject.asObservable();
  private loadingBehaviorSubject$ = new BehaviorSubject(false);
  loading$ = this.loadingBehaviorSubject$.asObservable();

  server: UtmServerType;

  private serversCache$: Observable<any> = this.utmServerService.query({ page: 0, size: 100 }).pipe(
    tap(response => {
      this.server = response.body[0];
    }),
    shareReplay(1),
    catchError(error => {
      this.utmToastService.showError(
        'Failed to fetch servers',
        'An error occurred while fetching server list.'
      );
      return of(null);
    })
  );

  modules$ = this.request$.pipe(
    filter((request): request is RequestModule => !!request),
    switchMap(request => this.fetchData(request))
  );

  constructor(
    private utmServerService: UtmServerService,
    private utmToastService: UtmToastService,
    private utmModulesService: UtmModulesService
  ) {}



  private fetchData(request: RequestModule): Observable<any[]> {
    return this.serversCache$.pipe(
      filter(response => !!response),
      tap(() => this.loadingBehaviorSubject$.next(true)),
      switchMap(response =>
        this.getModules({
          ...request,
          'serverId.equals': response.body[0].id
        })
      )
    );
  }

  private getModules(req: RequestModule): Observable<any[]> {
    return this.utmModulesService.getModules(req)
      .pipe(
        map(response => {
          response.body.map(m => {
            if (m.moduleName === this.utmModulesEnum.BITDEFENDER) {
              m.prettyName = m.prettyName + ' GravityZone';
            }
          });
          return response.body;
        }),
      catchError(error => {
        console.error(error);
        this.utmToastService.showError(
          'Failed to fetch modules',
          'An error occurred while fetching module data.'
        );
        return of([]);
      }),
      finalize(() => this.loadingBehaviorSubject$.next(false))
    );
  }

  loadModules(request: RequestModule): void {
    this.requestBehaviorSubject.next(request);
  }

  getModuleDetail(id: number) {
    this.loadingBehaviorSubject$.next(true);
    return this.utmModulesService.getModuleById(id)
      .pipe(
        catchError(() => {
          this.utmToastService.showError(
            'Failed to get module detail',
            'An error occurred while fetching module detail.'
          );
          return of(null);
        }),
        finalize(() => this.loadingBehaviorSubject$.next(false))
      );
  }
}

