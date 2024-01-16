import {Component, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {NavBehavior} from '../../shared/behaviors/nav.behavior';
import {SYSTEM_MENU_ICONS_PATH} from '../../shared/constants/menu_icons.constants';
import {UtmRunModeService} from '../../shared/services/active-modules/utm-run-mode.service';
import {UtmConfigSectionService} from '../../shared/services/config/utm-config-section.service';
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
export class AppModuleViewComponent implements OnInit {
  modules: UtmModuleType[];
  loading = true;
  setUpModule: UtmModulesEnum;
  utmModulesEnum = UtmModulesEnum;
  confValid = true;
  iconPath = SYSTEM_MENU_ICONS_PATH;
  active: UtmModuleType;
  module: UtmModuleType;
  category: any;
  categories: string[];
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

  constructor(public modalService: NgbModal,
              private activatedRoute: ActivatedRoute,
              private toastService: UtmToastService,
              private navBehavior: NavBehavior,
              private utmConfigSectionService: UtmConfigSectionService,
              private moduleRefreshBehavior: ModuleRefreshBehavior,
              private utmServerService: UtmServerService,
              private utmRunModeService: UtmRunModeService,
              private utmModulesService: UtmModulesService) {
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
    this.getServers();
  }

  getCategories() {
    this.utmModulesService.getModuleCategories({serverId: this.server.id, sort: 'moduleCategory,asc'})
      .subscribe(response => {
      this.categories = response.body.sort((a, b) => a > b ? 1 : -1);
    });
  }

  getServers() {
    this.utmServerService.query({page: 0, size: 100}).subscribe(response => {
      this.servers = response.body;
      this.server = this.servers[0];
      this.req['serverId.equals'] = this.server.id;
      this.getCategories();
      this.getModules();
    });
  }

  getModules() {
    this.utmModulesService.getModules(this.req).subscribe(response => {
      this.loading = false;
      this.modules = response.body;
    });
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
}
