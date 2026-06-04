import {AfterViewInit, ChangeDetectorRef, Component, OnDestroy, OnInit, QueryList, ViewChildren} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {NgbModal, NgbPopover} from '@ng-bootstrap/ng-bootstrap';
import {CompactType, GridsterConfig, GridsterItem, GridType} from 'angular-gridster2';
import {LocalStorageService} from 'ngx-webstorage';
import {UtmDashboardVisualizationType} from '../../../shared/chart/types/dashboard/utm-dashboard-visualization.type';
import {UtmDashboardType} from '../../../shared/chart/types/dashboard/utm-dashboard.type';
import {VisualizationType} from '../../../shared/chart/types/visualization.type';
import {ModalConfirmationComponent} from '../../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {CHART_BUILDER_ROUTE, TEXT_CHART_BUILDER_ROUTE} from '../../../shared/constants/app-routes.constant';
import {UTM_CHART_ICONS} from '../../../shared/constants/icons-chart.const';
import {ChartTypeEnum} from '../../../shared/enums/chart-type.enum';
import {RouteCallbackEnum} from '../../../shared/enums/route-callback.enum';
import {DashboardFilterType} from '../../../shared/types/filter/dashboard-filter.type';
// tslint:disable-next-line:max-line-length
import {
  VisualizationChangeNameComponent
} from '../../visualization/shared/components/visualization-change-name/visualization-change-name.component';
import {VisualizationService} from '../../visualization/shared/services/visualization.service';
import {DashboardFilterCreateComponent} from '../dashboard-filter-create/dashboard-filter-create.component';
import {DashboardSaveComponent} from '../dashboard-save/dashboard-save.component';
import {DASHBOARD_STORAGE_NAME} from '../shared/const/dashboard.const';
import {DashboardStatusEnum} from '../shared/enums/dashboard-status.enum';
import {IComponent, LayoutService} from '../shared/services/layout.service';
import {UtmDashboardVisualizationService} from '../shared/services/utm-dashboard-visualization.service';
import {UtmDashboardService} from '../shared/services/utm-dashboard.service';
import {DashboardDraftType} from '../shared/types/dashboard-draft.type';

@Component({
  selector: 'app-dashboard-create',
  templateUrl: './dashboard-create.component.html',
  styleUrls: ['./dashboard-create.component.scss']
})
export class DashboardCreateComponent implements OnInit, OnDestroy, AfterViewInit {
  loadingDashboard = false;
  addNewVisualization = true;
  // View all visualizations in doom and create a QueryList with result
  @ViewChildren('gridsterItem') gridsterItems: QueryList<GridsterItem>;

  filters: DashboardFilterType[] = [];
  chartIcons = UTM_CHART_ICONS;
  timeEnable: number[] = [];
  visDashboard: UtmDashboardVisualizationType[] = [];
  /**
   * Var to keep id of table relations to delete if editing dashboard
   */
  tempDelete: number[] = [];
  private dashboardId: any;
  dashboard: UtmDashboardType;

  public options: GridsterConfig = {
    gridType: GridType.ScrollVertical,
    compactType: CompactType.None,
    minCols: 30,
    // maxCols: 30,
    minRows: 1,
    minItemRows: 1,
    fixedRowHeight: 430,
    fixedColWidth: 500,
    // maxItemRows: 100,
    defaultItemCols: 1,
    defaultItemRows: 1,
    draggable: {
      enabled: true,
    },
    resizable: {
      enabled: true,
    },
    // itemChangeCallback: LayoutService.itemChange,
    itemResizeCallback: DashboardCreateComponent.itemResize,
    swap: true
  };
  activeTimeGridster: number;
  routeCallbackEnum = RouteCallbackEnum;
  editMode: boolean;
  chartType = ChartTypeEnum;

  static itemResize(item, itemComponent) {
  }

  constructor(private layoutService: LayoutService,
              public modalService: NgbModal,
              private router: Router,
              private cdr: ChangeDetectorRef,
              private activatedRoute: ActivatedRoute,
              private localStorage: LocalStorageService,
              private visualizationService: VisualizationService,
              private utmDashboardService: UtmDashboardService,
              private dashboardVisualizationService: UtmDashboardVisualizationService) {

  }


  get layout(): { grid: GridsterItem, visualization: VisualizationType } [] {
    return this.layoutService.layout;
  }

  get components(): IComponent[] {
    return this.layoutService.components;
  }

  ngAfterViewInit(): void {
    this.cdr.detectChanges();
  }

  ngOnDestroy(): void {
    this.layoutService.layout = [];
  }

  ngOnInit() {
    this.activatedRoute.queryParams.subscribe(params => {
      if (params.mode === DashboardStatusEnum.EDIT) {
        this.editMode = true;
        this.loadingDashboard = true;
        this.dashboardId = Number(params.dashboardId);
        this.addNewVisualization = false;
        this.getDashboardToEdit();
        this.addNewVisualization = false;
      } else if (params.mode === DashboardStatusEnum.DRAFT) {
        this.loadingDashboard = true;
        const draft = this.getDashboardDraft();
        this.layoutService.layout = draft ? draft.layout : [];
        this.filters = draft ? draft.filters : [];
        this.visDashboard = draft ? draft.visualization : [];
        this.timeEnable = draft ? draft.timeEnable : [];
        this.addNewVisualization = false;
        if (params.addVisualization) {
          this.visualizationService.find(Number.parseInt(params.addVisualization, 0)).subscribe(response => {
            // call method to select visualization to avoid conflict and code repetition
            this.loadingDashboard = false;
            this.onVisSelected([response.body], true);
          });
        } else {
          this.loadingDashboard = false;
        }
      }
    });
  }

  /**
   * After get callback from created visualization need to add to current layout dasboard
   * @param id Visualization ID
   */
  getVisualizationToAdd(id: number): Promise<boolean> {
    return new Promise<boolean>(resolve => {
      this.visualizationService.find(id).subscribe(response => {
        // call method to select visualization to avoid conflict and code repetition
        this.onVisSelected([response.body], true);
      });
      resolve(true);
    });
  }

  getDashboardToEdit() {
    const request = {
      page: 0,
      size: 10000,
      'idDashboard.equals': this.dashboardId,
      sort: 'order,asc'
    };
    this.dashboardVisualizationService.query(request).subscribe(visResponse => {
      this.visDashboard = visResponse.body;
      if (this.visDashboard.length > 0) {
        this.dashboard = this.visDashboard[0].dashboard;
        this.filters = this.dashboard.filters ? JSON.parse(this.dashboard.filters) : [];
        for (const vis of this.visDashboard) {
          if (vis.showTimeFilter) {
            this.timeEnable.push(vis.visualization.id);
          }
          const grid = JSON.parse(vis.gridInfo);
          this.layoutService.addItem(vis.visualization, grid);
        }
      } else {
        this.utmDashboardService.find(this.dashboardId).subscribe(response => {
          this.dashboard = response.body;
        });
      }
      this.loadingDashboard = false;
    });
  }

  saveDashboard() {
    // assign all visualization card un doom to var
    const saveModal = this.modalService.open(DashboardSaveComponent, {centered: true});
    saveModal.componentInstance.grid = this.orderLayout();
    saveModal.componentInstance.visTimeEnabled = this.timeEnable;
    saveModal.componentInstance.filters = JSON.stringify(this.filters);
    saveModal.componentInstance.layout = this.layout;
    saveModal.componentInstance.dashboard = this.dashboard;
    saveModal.componentInstance.delete = this.tempDelete;
    saveModal.componentInstance.dashboardCreated.subscribe(() => {
      this.layoutService.layout = [];
      this.deleteDashboardDraft();
    });
  }

  /**
   * sort array by y and x position
   */
  orderLayout() {
    let gr = this.gridsterItems.toArray();
    gr = gr.sort((a, b) => {
      return a.item.y - b.item.y || a.item.x - b.item.x;
    });
    return gr;
  }

  gridChange($event: any) {
  }

  /**
   * Add visualization to dashboard layout
   * @param $event Visualization array
   * @param override Attribute to determine if override visualization in layout
   */
  onVisSelected($event: VisualizationType[], override?: boolean) {
    console.log('EVENT:', $event);
    for (const vis of $event) {
      const indexVis = this.layout.findIndex(value => value.visualization.id === vis.id);
      vis.chartConfig = (typeof vis.chartConfig === 'string' && vis.chartType !== ChartTypeEnum.TEXT_CHART) ?
        JSON.parse(vis.chartConfig) : vis.chartConfig;
      this.timeEnable.push(vis.id);
      if (indexVis === -1) {
        this.layoutService.addItem(vis);
      }
      if (override && indexVis !== -1) {
        // when edit visualization need override current in layout in order to this, first
        // the visualization at the same position and then delete the previous one
        this.layoutService.addItem(vis, this.layout[indexVis].grid);
        this.deleteVisualization(this.layout[indexVis]);
      }
    }
  }

  editVisualization(vis: VisualizationType) {
    this.router.navigate([(vis.chartType === ChartTypeEnum.TEXT_CHART
      ? TEXT_CHART_BUILDER_ROUTE : CHART_BUILDER_ROUTE)],
      {
        queryParams: {
          type: vis.eventType,
          chart: vis.chartType,
          mode: 'edit',
          visualizationId: vis.id,
          callback: RouteCallbackEnum.DASHBOARD
        }
      }).then(() => {
    });
  }

  togglePopover(popover: NgbPopover, visualization: VisualizationType) {
    if (popover.isOpen()) {
      popover.close();
    } else {
      popover.open({visualization});
    }

  }

  closePopover(popover: NgbPopover) {
    popover.close();
  }

  openChangeNameModal(vis: VisualizationType) {
    const visChangeName = this.modalService.open(VisualizationChangeNameComponent, {centered: true});
    visChangeName.componentInstance.visualization = vis;
  }

  enableTimeFilter(visualization) {
    const index = this.timeEnable.findIndex(value => value === visualization.id);
    if (index === -1) {
      this.timeEnable.push(visualization.id);
    } else {
      this.timeEnable.splice(index, 1);
    }
  }

  /**
   * If editing add id to tempDelete array for delete after save dashboard
   * @param item Item grid to delete
   */
  deleteVisualization(item: { grid: GridsterItem; visualization: VisualizationType }) {
    if (this.dashboardId) {
      const index = this.visDashboard.findIndex(value => Number(value.idDashboard) === Number(this.dashboardId)
        && Number(value.idVisualization) === Number(item.visualization.id));
      const visDasId = this.visDashboard[index].id;
      this.tempDelete.push(visDasId);
      this.layoutService.deleteItem(item.grid);
    } else {
      this.layoutService.deleteItem(item.grid);
    }
  }

  onItemChange($event: Event) {
  }

  addVisualization() {
    this.addNewVisualization = true;
  }

  saveDashboardDraft() {
    const draft: DashboardDraftType = {
      filters: this.filters,
      layout: this.layout,
      timeEnable: this.timeEnable,
      visualization: this.visDashboard
    };
    this.localStorage.store(DASHBOARD_STORAGE_NAME, draft);
  }

  getDashboardDraft(): DashboardDraftType {
    return this.localStorage.retrieve(DASHBOARD_STORAGE_NAME);
  }

  deleteDashboardDraft() {
    this.localStorage.clear(DASHBOARD_STORAGE_NAME);
  }

  onCreateNewVisualization() {
    this.saveDashboardDraft();
  }

  cancelCreation() {
    this.deleteDashboardDraft();
    this.router.navigate(['/creator/dashboard/list']);
  }

  createFilter() {
    const modalFilter = this.modalService.open(DashboardFilterCreateComponent, {centered: true});
    modalFilter.componentInstance.filterAdd.subscribe(filter => {
      this.createOrReplaceFilter(filter);
    });
  }

  editFilter(filter: DashboardFilterType) {
    const modalFilter = this.modalService.open(DashboardFilterCreateComponent, {centered: true});
    modalFilter.componentInstance.filter = filter;
    modalFilter.componentInstance.filterAdd.subscribe(filterEdit => {
      this.createOrReplaceFilter(filterEdit);
    });
  }

  onAction($event: { action: 'edit' | 'delete'; filter: DashboardFilterType }) {
    if ($event.action === 'delete') {
      this.deleteFilter($event.filter);
    } else {
      this.editFilter($event.filter);
    }
  }

  createOrReplaceFilter(filter: DashboardFilterType) {
    const index = this.filters.findIndex(value => value.id === filter.id);
    if (index === -1) {
      this.filters.push(filter);
    } else {
      this.filters[index] = filter;
    }
  }

  deleteFilter(filter: DashboardFilterType) {
    const deleteFilter = this.modalService.open(ModalConfirmationComponent, {centered: true});
    deleteFilter.componentInstance.header = 'Delete dashboard filter';
    deleteFilter.componentInstance.message = 'Are you sure that you want to delete this filter?';
    deleteFilter.componentInstance.confirmBtnText = 'Delete';
    deleteFilter.componentInstance.confirmBtnIcon = 'icon-filter';
    deleteFilter.componentInstance.confirmBtnType = 'delete';
    deleteFilter.result.then(() => {
      const index = this.filters.indexOf(filter);
      this.filters.splice(index, 1);
    });
  }
}
