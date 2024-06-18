import {HttpResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {combineLatest, Observable, throwError} from 'rxjs';
import {catchError, map, switchMap, tap} from 'rxjs/operators';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {SYSTEM_MENU_ICONS_PATH} from '../../shared/constants/menu_icons.constants';
import {ModuleRefreshBehavior} from '../shared/behavior/module-refresh.behavior';
import {UtmModulesEnum} from '../shared/enum/utm-module.enum';
import {UtmModulesService} from '../shared/services/utm-modules.service';
import {UtmServerService} from '../shared/services/utm-server.service';
import {UtmModuleType} from '../shared/type/utm-module.type';
import {UtmServerType} from '../shared/type/utm-server.type';

interface ModuleData {
  modules: UtmModuleType[];
  categories: string[];
}

@Component({
  selector: 'app-app-module-view',
  templateUrl: './app-module-view.component.html',
  styleUrls: ['./app-module-view.component.scss']
})
export class AppModuleViewComponent implements OnInit {
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
  servers: UtmServerType[];

  moduleData$: Observable<ModuleData>;

  constructor(public modalService: NgbModal,
              private activatedRoute: ActivatedRoute,
              private moduleRefreshBehavior: ModuleRefreshBehavior,
              private utmServerService: UtmServerService,
              private utmModulesService: UtmModulesService,
              private utmToastService: UtmToastService) {

    this.activatedRoute.queryParams.subscribe(params => {
      if (params) {
        this.req['moduleName.equals'] = params.setUp;
      }
    });
    this.moduleRefreshBehavior.$moduleChange.subscribe(refresh => {
      this.getModules();
    });
  }

  ngOnInit() {
  /* this.modules$ = this.getDataModule();
   this.categories$ = this.getCategories();*/

    this.moduleData$ = this.getDataModule();
  }

  getCategories() {
    return this.utmModulesService.getModuleCategories(
        {
              serverId: this.server.id,
              sort: 'moduleCategory,asc'
        })
        .pipe(
            tap(() => this.loading = !this.loading),
            map( res => {
              return res.body.sort((a, b) => a > b ? 1 : -1);
            }));
  }

 /* getServers() {
    this.utmServerService.query({page: 0, size: 100}).subscribe(response => {
      this.servers = response.body;
      console.log(this.servers);
      this.server = this.servers[0];
      this.req['serverId.equals'] = this.server.id;
      this.getCategories();
      this.getModules();
    });
  }*/

  getDataModule(): Observable<ModuleData> {
    return this.utmServerService.query({page: 0, size: 100})
        .pipe(
            map((resp: HttpResponse<UtmServerType[]>) => resp.body),
            switchMap(servers => {
              this.server = servers[0];
              this.req['serverId.equals'] = servers[0].id;
              return combineLatest(this.getModules(), this.getCategories())
                  .pipe(
                      map(([modules, categories]) => {
                        return {
                          modules,
                          categories
                        };
                      }),
                  catchError(error => {
                    console.log(error);
                    this.utmToastService.showError('Failed to list modules',
                        'An error occurred while fetching module data.');
                    return [];
                  })
              );
            })
        );
  }

  /*getModules() {
    this.utmModulesService.getModules(this.req)
        .pipe(
            map( response => {
              response.body.map(m => {
                if (m.moduleName === this.utmModulesEnum.BITDEFENDER){
                   m.prettyName = m.prettyName + ' GravityZone';
                }
              });

              return response.body;
            })
        )
        .subscribe(modules => {
      this.loading = false;
      this.modules = modules;
    });
  }*/

  getModules(): Observable<UtmModuleType[]> {
    return this.utmModulesService.getModules(this.req)
        .pipe(
            map( response => {
              response.body.map(m => {
                if (m.moduleName === this.utmModulesEnum.BITDEFENDER){
                   m.prettyName = m.prettyName + ' GravityZone';
                }
              });

              return response.body;
            }),
            tap ((modules) => {
              this.loading = false;
            })
        );
  }

  showModule($event: UtmModuleType) {
    this.module = $event;
  }

  filterByCategory($event: any) {
    this.req['moduleCategory.equals'] = $event;
    this.getModules();

  }

  onSearch($event: string) {
    this.req.page = 0;
    this.req['prettyName.contains'] = $event;
    this.getModules();
  }

  filterByServer($event: any) {
    this.req['serverId.equals'] = $event.id;
    this.req['moduleCategory.equals'] = null;
    this.server = $event;
    this.category = undefined;
    this.getCategories();
    this.getModules();
  }

  trackByFn(index: number, module: UtmModuleType): any {
    return module.id;
  }
}
