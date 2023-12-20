import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {Router} from '@angular/router';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {GridsterItem} from 'angular-gridster2';
import {Observable} from 'rxjs';
import {UserService} from '../../../core/user/user.service';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {NavBehavior} from '../../../shared/behaviors/nav.behavior';
import {UtmDashboardVisualizationType} from '../../../shared/chart/types/dashboard/utm-dashboard-visualization.type';
import {UtmDashboardType} from '../../../shared/chart/types/dashboard/utm-dashboard.type';
import {VisualizationType} from '../../../shared/chart/types/visualization.type';
import {MenuCreateComponent} from '../../../shared/components/utm/util/menu-create/menu-create.component';
import {UTM_CHART_ICONS} from '../../../shared/constants/icons-chart.const';
import {TIME_DASHBOARD_REFRESH} from '../../../shared/constants/time-refresh.const';
import {Operator} from '../../../shared/enums/operator.enum';
import {MenuService} from '../../../shared/services/menu/menu.service';
import {ElasticFiltersType} from '../../../shared/types/filter/elastic-filters.type';
import {Menu} from '../../../shared/types/menu/menu.model';
import {QueryType} from '../../../shared/types/query-type';
import {InputClassResolve} from '../../../shared/util/input-class-resolve';
import {UtmDashboardVisualizationService} from '../shared/services/utm-dashboard-visualization.service';
import {UtmDashboardService} from '../shared/services/utm-dashboard.service';
import {buildDashboardUrl} from '../shared/util/get-menu-url';

@Component({
  selector: 'app-dashboard-save',
  templateUrl: './dashboard-save.component.html',
  styleUrls: ['./dashboard-save.component.scss']
})
export class DashboardSaveComponent implements OnInit {
  @Input() grid: GridsterItem[];
  @Input() dashboard: UtmDashboardType;
  @Input() filters: ElasticFiltersType;
  @Input() visTimeEnabled: number[];
  @Input() delete: number[];
  @Input() layout: { grid: GridsterItem, visualization: VisualizationType } [];
  @Output() dashboardCreated = new EventEmitter<string>();
  dashSaveForm: FormGroup;
  creating = false;
  saveMode: boolean;
  refreshTime = TIME_DASHBOARD_REFRESH;
  visualizationToSave = '';
  saveToMenu = false;
  menuList: Menu[] = [];
  menu: Menu;
  subMenuName: string;
  authorities: string[];
  authority: string[];
  menuId: number;
  menuActive = true;
  private idMenu: number;
  private asNewMenu: boolean;

  constructor(public activeModal: NgbActiveModal,
              public inputClassResolve: InputClassResolve,
              private fb: FormBuilder,
              private utmToastService: UtmToastService,
              private menuService: MenuService,
              private menuBehavior: NavBehavior,
              private modalService: NgbModal,
              private utmDashboardVisualizationService: UtmDashboardVisualizationService,
              private utmDashboardService: UtmDashboardService,
              private router: Router,
              private userService: UserService) {
  }

  ngOnInit() {
    this.initFormSaveVis();
    if (this.dashboard) {
      this.dashSaveForm.patchValue(this.dashboard);
      this.dashSaveForm.get('autoRefresh').setValue(this.dashboard.refreshTime !== null);
      this.getMenuEdit();
    }
    this.dashSaveForm.get('filters').setValue(this.filters);
    this.getMenuList();
    this.getAuthoritiesList();
    this.dashSaveForm.get('name').valueChanges.subscribe(value => this.subMenuName = value);
  }

  initFormSaveVis() {
    this.dashSaveForm = this.fb.group(
      {
        name: ['', [Validators.required]],
        description: [''],
        createdDate: [new Date()],
        filters: [],
        id: [],
        modifiedDate: [new Date()],
        userCreated: [1],
        userModified: [1],
        autoRefresh: [false],
        refreshTime: [],
        systemOwner: []
      }
    );
  }

  getAuthoritiesList() {
    this.userService.authorities().subscribe(result => {
      if (result) {
        this.authorities = result;
      }
    });
  }

  getMenuList() {
    const query = new QueryType();
    query.add('parentId', false, Operator.specified);
    query.add('size', 10000);
    query.add('url', false, Operator.specified);
    query.add('modulesNameShort', false, Operator.specified);
    this.menuService.query(query).subscribe(menus => {
      this.menuList = menus.body;
    });
  }

  /**
   * Set menu var if dashboard is already in navigation abr
   */
  getMenuEdit() {
    const query = new QueryType();
    query.add('dashboardId', this.dashboard.id, Operator.equals);
    this.menuService.query(query).subscribe(menus => {
      if (menus.body.length > 0) {
        this.idMenu = menus.body[0].id;
        this.authority = menus.body[0].authorities;
        this.subMenuName = menus.body[0].name;
        this.menuId = menus.body[0].parentId;
        this.menuActive = menus.body[0].menuActive;
        this.saveToMenu = true;
        this.asNewMenu = false;
      } else {
        this.asNewMenu = true;
      }
    });
  }

  saveDashboard() {
    this.creating = true;
    if (this.dashboard && !this.saveMode) {
      this.editDashboard();
    } else {
      this.createDashboard();
    }
  }

  createDashboard() {
    this.dashSaveForm.get('id').setValue(null);
    this.utmDashboardService.create(this.dashSaveForm.value).subscribe(dashboard => {
      if (this.saveToMenu) {
        this.saveDashboardToMenu(dashboard.body);
      }
      this.creatingVisualizationsRecords(dashboard.body).then(created => {
        this.utmToastService.showSuccessBottom('Dashboard created successfully');
        this.dashboardCreated.emit('created');
        this.menuBehavior.$nav.next(true);
        this.router.navigate(['/creator/dashboard/list']);
        this.activeModal.close();
      });
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error creating dashboard',
        'Error while trying to create new dashboard');
    });
  }

  /**
   * Check in database that relations do not exist to create new one
   * @param dashboard Dashboard to save
   */
  creatingVisualizationsRecords(dashboard: UtmDashboardType): Promise<string> {
    return new Promise<string>((resolve, reject) => {
      // for (const item of this.grid) {
      const promises = [];
      for (let j = 0; j < this.grid.length; j++) {
        const item = this.grid[j];
        const query = {
          'idDashboard.equals': dashboard.id,
          'idVisualization.equals': this.getVisualizationId(item),
        };
        this.visualizationToSave = this.getVisualizationName(item);
        const dasVis: UtmDashboardVisualizationType = {
          height: item.height,
          idDashboard: dashboard.id,
          idVisualization: this.getVisualizationId(item),
          left: item.item.x,
          order: j,
          top: item.item.y,
          width: item.width,
          gridInfo: JSON.stringify(item.item),
          showTimeFilter: this.visTimeEnabled.includes(this.getVisualizationId(item))
        };
        promises.push(this.insertUpdateVisualizationAsync(query, dasVis));
      }
      Promise.all(promises)
        .then((results) => {
          resolve('created');
        })
        .catch((e) => {
          // Handle errors here
        });
    });
  }

  insertUpdateVisualizationAsync(query, dasVis) {
    return new Promise((resolve) => {
      this.utmDashboardVisualizationService.query(query).subscribe(response => {
        console.log('Creating relation between', dasVis.idVisualization, dasVis.idDashboard, response.body.length);
        if (response.body.length === 0) {
          this.utmDashboardVisualizationService.create(dasVis).subscribe(res => {
            resolve(res.body.idDashboard);
          });
        } else {
          dasVis.id = response.body[0].id;
          this.utmDashboardVisualizationService.update(dasVis).subscribe(res => {
            resolve(res.body.idDashboard);
          });
        }
      }, error => {
        console.log('Error trying to get relation between dashboard and visualization');
      });
    });
  }

  editDashboard() {
    this.utmDashboardService.update(this.dashSaveForm.value).subscribe(response => {
      if (this.saveToMenu) {
        if (!this.asNewMenu) {
          this.editDashboardToMenu(response.body);
        } else {
          this.saveDashboardToMenu(response.body);
        }
      }
      this.creatingVisualizationsRecords(response.body).then(created => {
        this.deleteVisualizationRelation().subscribe(value => {
          this.utmToastService.showSuccessBottom('Dashboard edited successfully');
          this.dashboardCreated.emit('created');
          this.menuBehavior.$nav.next(true);
          this.router.navigate(['/creator/dashboard/list']);
          this.activeModal.close();
        });
      });
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error editing dashboard',
        error1.error.statusText);
    });
  }


  chartIconResolver(chartType: string) {
    return UTM_CHART_ICONS[chartType];
  }

  saveAsNew($event: boolean) {
    this.saveMode = $event;
    if ($event) {
      this.dashSaveForm.get('name').setValue(this.dashboard.name + '-Clone');
    } else {
      this.dashSaveForm.get('name').setValue(this.dashboard.name);
    }
  }

  saveDashboardToMenu(dashboard: UtmDashboardType) {
    const menu: Menu = {
      authorities: this.authority,
      name: this.subMenuName,
      parentId: this.menuId,
      type: 3,
      menuActive: this.menuActive,
      url: this.getDashboardUrl(dashboard),
      menuAction: false,
      dashboardId: dashboard.id
    };
    this.menuService.create(menu).subscribe(value => {
    });
  }

  getDashboardUrl(dashboard: UtmDashboardType): string {
    return buildDashboardUrl(dashboard);
  }

  addMenu() {
    const modalRef = this.modalService.open(MenuCreateComponent,
      {backdrop: 'static', centered: true});
    modalRef.componentInstance.level = 1;
    modalRef.componentInstance.isCreate = true;
    modalRef.componentInstance.menuCreated.subscribe(menu => {
      this.getMenuList();
      this.menuId = menu.id;
    });
  }

  private getVisualizationId(grid: GridsterItem): number {
    const index = this.layout.findIndex(value => value.grid.id === grid.item.id);
    return this.layout[index].visualization.id;
  }

  private getVisualizationName(grid: GridsterItem) {
    const index = this.layout.findIndex(value => value.grid.id === grid.item.id);
    return this.layout[index].visualization.name;
  }

  private editDashboardToMenu(dashboard: UtmDashboardType) {
    const menu: Menu = {
      id: this.idMenu,
      authorities: this.authority,
      name: this.subMenuName,
      parentId: this.menuId,
      type: 3,
      url: this.getDashboardUrl(dashboard),
      menuActive: this.menuActive,
      dashboardId: dashboard.id
    };
    this.menuService.update(menu).subscribe(value => {
    });
  }

  private deleteVisualizationRelation(): Observable<boolean> {
    return new Observable<boolean>(subscriber => {
      for (const id of this.delete) {
        this.utmDashboardVisualizationService.delete(id).subscribe(value => {
        });
      }
      subscriber.next(true);
    });

  }

  saveToMenuToggle($event: boolean) {
    this.saveToMenu = $event;
    this.subMenuName = this.dashSaveForm.get('name').value;
  }
}
