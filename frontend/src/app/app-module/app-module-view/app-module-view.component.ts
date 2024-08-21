import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {Observable, of, Subject} from 'rxjs';
import {catchError, filter, map, takeUntil, tap} from 'rxjs/operators';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {SYSTEM_MENU_ICONS_PATH} from '../../shared/constants/menu_icons.constants';
import {ModuleResolverService} from '../services/module.resolver.service';
import {ModuleRefreshBehavior} from '../shared/behavior/module-refresh.behavior';
import {UtmModulesEnum} from '../shared/enum/utm-module.enum';
import {UtmModulesService} from '../shared/services/utm-modules.service';
import {UtmServerService} from '../shared/services/utm-server.service';
import {UtmModuleType} from '../shared/type/utm-module.type';
import {UtmServerType} from '../shared/type/utm-server.type';

@Component({
  selector: 'app-app-module-view',
  templateUrl: './app-module-view.component.html',
  styleUrls: ['./app-module-view.component.scss']
})
export class AppModuleViewComponent implements OnInit, OnDestroy {
  modules: UtmModuleType[];
  modules$: Observable<UtmModuleType[]>;
  loading = true;
  setUpModule: UtmModulesEnum;
  utmModulesEnum = UtmModulesEnum;
  confValid = true;
  iconPath = SYSTEM_MENU_ICONS_PATH;
  active: UtmModuleType;
  module: UtmModuleType;
  category: any;
  categories: string[];
  categories$: Observable<string[]>;
  req = {
    'moduleCategory.equals': null,
    'prettyName.contains': null,
    'serverId.equals': null,
    sort: 'moduleCategory,asc',
    'moduleName.equals': null,
    page: 0,
    size: 100,
  };
  server: UtmServerType;
  destroy$ = new Subject<void>();

  constructor(public modalService: NgbModal,
              private activatedRoute: ActivatedRoute,
              private moduleRefreshBehavior: ModuleRefreshBehavior,
              private utmModulesService: UtmModulesService,
              private utmToastService: UtmToastService,
              private moduleResolver: ModuleResolverService) {

    this.activatedRoute.queryParams.subscribe(params => {
      if (params) {
        this.req['moduleName.equals'] = params.setUp;
      }
    });

    this.moduleRefreshBehavior.$moduleChange
      .pipe(
        takeUntil(this.destroy$),
        filter(value => !!value))
      .subscribe(refresh => {
      this.refreshModules();
    });
  }

  ngOnInit() {
    this.modules$ = this.activatedRoute.data
      .pipe(
        map(data => data.response),
        tap(() => {
          this.server = this.moduleResolver.server;
          this.req['serverId.equals'] = this.server.id;
          this.getCategories();
        })
      );
  }

  getCategories() {
      this.categories$ = this.utmModulesService
          .getModuleCategories({serverId: this.server.id, sort: 'moduleCategory,asc'})
            .pipe(
                tap(() => this.loading = !this.loading),
                map( res => {
                  return res.body ? res.body.sort((a, b) => a > b ? 1 : -1) : [];
                }),
                catchError(error => {
                    console.log(error);
                    this.utmToastService.showError('Failed to fetch categories',
                        'An error occurred while fetching module data.');
                    return of([]);
                })
            );
  }

  refreshModules() {
    this.loading = true;
    this.modules$ = this.moduleResolver.getModules(this.req);
  }

  showModule($event: UtmModuleType) {
    this.module = $event;
  }

  filterByCategory($event: any) {
    console.log('filter');
    this.req['moduleCategory.equals'] = $event;
    this.refreshModules();

  }

  onSearch($event: string) {
    console.log('search');
    this.req.page = 0;
    this.req['prettyName.contains'] = $event;
    this.refreshModules();
  }

  filterByServer($event: any) {
    this.req['serverId.equals'] = $event.id;
    this.req['moduleCategory.equals'] = null;
    this.server = $event;
    this.category = undefined;
    this.getCategories();
    this.refreshModules();
  }

  trackByFn(index: number, module: UtmModuleType): any {
    return module.id;
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
