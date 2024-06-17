import {Component, OnDestroy, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {ResizeEvent} from 'angular-resizable-element';
import {of} from 'rxjs';
import {catchError, finalize, map} from 'rxjs/operators';
import {UtmModuleCollectorService} from '../../app-module/shared/services/utm-module-collector.service';
import {UtmModuleCollectorType} from '../../app-module/shared/type/utm-module-collector.type';
import {UtmModuleGroupConfType} from '../../app-module/shared/type/utm-module-group-conf.type';
import {AccountService} from '../../core/auth/account.service';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {ModalConfirmationComponent} from '../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {ITEMS_PER_PAGE} from '../../shared/constants/pagination.constants';
import {SortEvent} from '../../shared/directives/sortable/type/sort-event';
import {IncidentOriginTypeEnum} from '../../shared/enums/incident-response/incident-origin-type.enum';
import {IncidentCommandType} from '../../shared/types/incident/incident-command.type';
import {UtmFieldType} from '../../shared/types/table/utm-field.type';
import {TimeFilterType} from '../../shared/types/time-filter.type';
import {calcTableDimension} from '../../shared/util/screen.util';
import {AssetFiltersBehavior} from '../shared/behavior/asset-filters.behavior';
import {AssetSaveReportComponent} from '../shared/components/asset-save-report/asset-save-report.component';
import {ASSETS_FIELDS, COLLECTORS_FIELDS_FILTERS} from '../shared/const/asset-field.const';
import {STATICS_FILTERS} from '../shared/const/filter-const';
import {AssetFieldEnum} from '../shared/enums/asset-field.enum';
import {CollectorFieldFilterEnum} from '../shared/enums/collector-field-filter.enum';
import {UtmNetScanService} from '../shared/services/utm-net-scan.service';
import {AssetFilterType} from '../shared/types/asset-filter.type';
import {UtmDataInputStatus} from '../shared/types/data-source-input.type';
import {NetScanType} from '../shared/types/net-scan.type';
import {SourceDataTypeConfigComponent} from '../source-data-type-config/source-data-type-config.component';

@Component({
  selector: 'app-collectors-view',
  templateUrl: './collectors-view.component.html',
  styleUrls: ['./collectors-view.component.scss']
})
export class CollectorsViewComponent implements OnInit, OnDestroy {
  assets: NetScanType[] = [];
  collectors: UtmModuleCollectorType[] = [];
  // defaultTime: ElasticFilterDefaultTime = new ElasticFilterDefaultTime('now-30d', 'now');
  pageWidth = window.innerWidth;
  filterWidth: number;
  tableWidth: number;
  sortEvent: any;
  totalItems: any;
  page = 0;
  loading = true;
  itemsPerPage = ITEMS_PER_PAGE;
  viewAssetDetail: UtmModuleCollectorType;
  sortBy = AssetFieldEnum.ASSET_ID + ',asc';
  assetsFields: UtmFieldType[] = ASSETS_FIELDS;
  checkbox: any;
  assetFieldEnum = AssetFieldEnum;
  fieldFilters = COLLECTORS_FIELDS_FILTERS;
  requestParam: AssetFilterType = {
    alive: null,
    discoveredEndDate: null,
    discoveredInitDate: null,
    openPorts: null,
    os: null,
    page: 0,
    severity: null,
    probe: null,
    alias: null,
    size: ITEMS_PER_PAGE,
    sort: 'id,desc',
    status: null,
    type: null,
    groups: null
  };
  assetsSelected: number[] = [];
  interval: any;
  deleting: string[] = [];
  reasonRun: IncidentCommandType;
  agent: string;
  configs: UtmModuleGroupConfType[] = [];
  showDetail = false;

  constructor(private utmNetScanService: UtmNetScanService,
              private modalService: NgbModal,
              private utmToastService: UtmToastService,
              private accountService: AccountService,
              private assetFiltersBehavior: AssetFiltersBehavior,
              private collectorService: UtmModuleCollectorService,
              private toast: UtmToastService) {
  }

  // Init get asset on time filter component trigger


  ngOnInit() {
    this.setInitialWidth();
    this.getCollectors();
    this.starInterval();
    this.accountService.identity().then(account => {
      this.reasonRun = {
        command: '',
        reason: '',
        originId: account.login,
        originType: IncidentOriginTypeEnum.USER_EXECUTION
      };
    });

  }
  setInitialWidth() {
    const dimensions = calcTableDimension(this.pageWidth);
    this.tableWidth = dimensions.tableWidth;
    this.filterWidth = dimensions.filterWidth;
  }

  loadPage($event: number) {
    this.requestParam.page = $event - 1;
    this.getCollectors();
  }
  getCollectors() {
    this.collectorService.queryFilter(this.requestParam)
        .pipe(
            catchError(error => {
              this.toast.showError('Error listing collectors',
                  'An error occurred while trying to list the collectors. Please try again.');
              return of(null);
            }),
            finalize(() => this.loading = false)
        )
        .subscribe(response => {
          this.totalItems = Number(response.headers.get('X-Total-Count'));
          this.collectors = response.body;
        });
  }

  onItemsPerPageChange($event: number) {
    this.itemsPerPage = $event;
    this.requestParam.size = $event;
    this.getCollectors();
  }

  onTimeFilterChange($event: TimeFilterType) {
    this.requestParam.discoveredInitDate = $event.timeFrom;
    this.requestParam.discoveredEndDate = $event.timeTo;
    this.getCollectors();
  }

  onResize($event: ResizeEvent) {
    if ($event.rectangle.width >= 250) {
      this.tableWidth = (this.pageWidth - $event.rectangle.width - 51);
      this.filterWidth = $event.rectangle.width;
    }
  }

  saveReport() {
    const reportModal = this.modalService.open(AssetSaveReportComponent, {centered: true});
    reportModal.componentInstance.assetFilters = this.requestParam;
  }

  onSortBy($event: SortEvent) {
    this.requestParam.sort = $event.column + ',' + $event.direction;
    this.getCollectors();
  }

  toggleCheck() {
    this.checkbox = !this.checkbox;
    if (!this.checkbox) {
      this.assetsSelected = [];
    } else {
      this.assetsSelected = this.assets.map(value => value.id);
    }
  }

  addToSelected(event: Event, asset: UtmModuleCollectorType) {
    event.stopPropagation();
    const index = this.assetsSelected.findIndex(value => value === asset.id);
    if (index === -1) {
      this.assetsSelected.push(asset.id);
    } else {
      this.assetsSelected.splice(index, 1);
    }
  }

  isSelected(asset: UtmModuleCollectorType): boolean {
    return this.assetsSelected.findIndex(value => value === asset.id) !== -1;
  }

  resetAllFilters() {
    for (const key of Object.keys(this.requestParam)) {
      if (!STATICS_FILTERS.includes(key)) {
        this.requestParam[key] = null;
      }
    }
    this.assetFiltersBehavior.$assetFilter.next(this.requestParam);
    this.getCollectors();
  }

  onFilterChange($event: { prop: CollectorFieldFilterEnum, values: any }) {
    switch ($event.prop) {
      case CollectorFieldFilterEnum.COLLECTOR_GROUP:
        this.requestParam.groups = $event.values.length > 0 ? $event.values : null;
        break;
    }
    this.assetFiltersBehavior.$assetAppliedFilter.next(this.requestParam);
    this.assetFiltersBehavior.$assetFilter.next(this.requestParam);
    this.getCollectors();
  }

  onSearch($event: string) {
    this.requestParam.assetIpMacName = $event;
    this.requestParam.page = 0;
    this.getCollectors();
  }
  getCollectorSource(asset: UtmModuleCollectorType) {
    if (asset.hostname && asset.ip) {
      return asset.hostname + ' (' + asset.ip + ')';
    } else if (asset.hostname) {
      return asset.hostname;
    } else if (asset.ip) {
      return asset.ip;
    } else {
      return 'Unknown source';
    }
  }

  viwAgentDetail(event: Event, asset: UtmModuleCollectorType) {
    event.stopPropagation();
    this.viewAssetDetail = asset;
    this.agent = asset.hostname;
    this.collectorService.groups(asset.id.toString())
        .subscribe(res => {
          this.configs = [];
          res.body.forEach((item: { moduleGroupConfigurations: any; }) => {
            const configurations = item.moduleGroupConfigurations;
            this.configs.push(...configurations);
          });

          this.showDetail = true;
        });
  }

  getHostnames() {
    return this.configs.filter(conf => conf.confName === 'Hostname')
        .map(conf => conf.confValue);
  }

  isSourceConnected(asset: NetScanType, source: UtmDataInputStatus): boolean {
    if (asset.agent && !asset.assetAlive) {
      return false;
    } else {
      return !source.down;
    }
  }

  showDataTypeModal() {
    const modalSource = this.modalService.open(SourceDataTypeConfigComponent, {centered: true, size: 'lg'});
    modalSource.componentInstance.refreshDataInput.subscribe(() => {
      this.getCollectors();
    });
  }
  closeDetail() {
    this.showDetail = false;
  }

  stopInterval(event: boolean) {
    if (event) {
      clearInterval(this.interval);
      this.interval = null;
    } else {
      this.starInterval();
    }
  }

  starInterval() {
    if (!this.interval) {
      this.interval = setInterval(() => {
        this.getCollectors();
      }, 10000);
    }
  }

  deleteCollector($event: MouseEvent, collector: UtmModuleCollectorType) {
    event.stopPropagation();
    const deleteModalRef = this.modalService.open(ModalConfirmationComponent, {centered: true});
    deleteModalRef.componentInstance.header = 'Delete collector';
    deleteModalRef.componentInstance.message = 'Are you sure that you want to delete this collector?';
    deleteModalRef.componentInstance.confirmBtnText = 'Delete';
    deleteModalRef.componentInstance.confirmBtnIcon = 'icon-display';
    deleteModalRef.componentInstance.confirmBtnType = 'delete';
    deleteModalRef.result.then(() => {
      this.delete(collector);
    });
  }

  delete(collector: UtmModuleCollectorType) {
    this.loading = true;
    this.collectorService.deleteCollector(collector.id)
            .pipe(
                catchError(error => {
                  this.toast.showError('Error deleting collector',
                      'An error occurred while trying to delete the collector. Please try again.');

                  return of(null);
                }),
                finalize(() => this.loading = false)
            )
            .subscribe(() => {
              this.getCollectors();
              this.toast.showSuccessBottom(`${collector.hostname} deleted successfully`);
            });
  }

  ngOnDestroy(): void {
    this.stopInterval(true);
    this.assetFiltersBehavior.$assetFilter.next(null);
  }

}
