import {ChangeDetectionStrategy, Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {Observable, of, Subject} from 'rxjs';
import {catchError, filter, finalize, map, takeUntil, tap} from 'rxjs/operators';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {SYSTEM_MENU_ICONS_PATH} from '../../shared/constants/menu_icons.constants';
import {ModuleService} from '../services/module.service';
import {ModuleRefreshBehavior} from '../shared/behavior/module-refresh.behavior';
import {UtmModulesEnum} from '../shared/enum/utm-module.enum';
import {UtmModulesService} from '../shared/services/utm-modules.service';
import {UtmServerService} from '../shared/services/utm-server.service';
import {UtmModuleType} from '../shared/type/utm-module.type';
import {UtmServerType} from '../shared/type/utm-server.type';

@Component({
  selector: 'app-app-module-view',
  templateUrl: './app-module-view.component.html',
  styleUrls: ['./app-module-view.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class AppModuleViewComponent implements OnInit, OnDestroy {
  modules: UtmModuleType[];
  modules$: Observable<UtmModuleType[]>;
  loading = false;
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
              public moduleResolver: ModuleService) {
  }

  ngOnInit() {
    /*this.modules$ = this.activatedRoute.data
      .pipe(
        map(data => data.response),
        tap(() => {
          this.server = this.moduleResolver.server;
          this.req['serverId.equals'] = this.server.id;
          this.getCategories();
        })
      );*/

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
        this.resetRequest();
        this.moduleResolver.loadModules({
          ...this.req
        });
      });

    this.refreshModules();
  }

  getCategories() {
      this.categories$ = this.utmModulesService
          .getModuleCategories({serverId: this.moduleResolver.server.id, sort: 'moduleCategory,asc'})
            .pipe(
                tap(() => this.loading = true),
                map( res => {
                  return res.body ? res.body.sort((a, b) => a > b ? 1 : -1) : [];
                }),
                catchError(error => {
                    this.utmToastService.showError('Failed to fetch categories',
                        'An error occurred while fetching module data.');
                    return of([]);
                }),
              finalize(() => this.loading = false),
            );
  }

  refreshModules() {
    this.moduleResolver.loadModules({
      ...this.req
    });
  }

  showModule($event: UtmModuleType) {
    this.moduleResolver.getModuleDetail($event.id)
      .subscribe((module: UtmModuleType) => {
        this.module = module;
        this.loading = false;
      });
  }

  filterByCategory($event: any) {
    this.req['moduleCategory.equals'] = $event;
    this.refreshModules();

  }

  onSearch($event: string) {
    this.req.page = 0;
    this.req['prettyName.contains'] = $event;
    this.refreshModules();
  }

  trackByFn(index: number, module: UtmModuleType): any {
    return module.id;
  }

  resetRequest() {
    this.req = {
      'moduleCategory.equals': null,
      'prettyName.contains': null,
      'serverId.equals': null,
      sort: 'moduleCategory,asc',
      'moduleName.equals': null,
      page: 0,
      size: 100,
    };
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
