import {Component, OnDestroy, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {ResizeEvent} from 'angular-resizable-element';
import * as moment from 'moment';
import {NgxSpinnerService} from 'ngx-spinner';
import {AccountService} from '../../core/auth/account.service';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {ModalConfirmationComponent} from '../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {ALERT_SENSOR_FIELD} from '../../shared/constants/alert/alert-field.constant';
import {ITEMS_PER_PAGE} from '../../shared/constants/pagination.constants';
import {SortEvent} from '../../shared/directives/sortable/type/sort-event';
import {ChartValueSeparator} from '../../shared/enums/chart-value-separator';
import {ElasticOperatorsEnum} from '../../shared/enums/elastic-operators.enum';
import {IncidentOriginTypeEnum} from '../../shared/enums/incident-response/incident-origin-type.enum';
import {IncidentCommandType} from '../../shared/types/incident/incident-command.type';
import {UtmFieldType} from '../../shared/types/table/utm-field.type';
import {TimeFilterType} from '../../shared/types/time-filter.type';
import {calcTableDimension} from '../../shared/util/screen.util';
import {AssetFiltersBehavior} from '../shared/behavior/asset-filters.behavior';
import {AssetSaveReportComponent} from '../shared/components/asset-save-report/asset-save-report.component';
import {ASSETS_FIELDS, ASSETS_FIELDS_FILTERS} from '../shared/const/asset-field.const';
import {STATICS_FILTERS} from '../shared/const/filter-const';
import {AssetFieldFilterEnum} from '../shared/enums/asset-field-filter.enum';
import {AssetFieldEnum} from '../shared/enums/asset-field.enum';
import {DataSourceInputService} from '../shared/services/data-source-input.service';
import {UtmNetScanService} from '../shared/services/utm-net-scan.service';
import {AssetFilterType} from '../shared/types/asset-filter.type';
import {UtmDataInputStatus} from '../shared/types/data-source-input.type';
import {NetScanType} from '../shared/types/net-scan.type';
import {SourceDataTypeConfigComponent} from '../source-data-type-config/source-data-type-config.component';

@Component({
  selector: 'app-assets-view',
  templateUrl: './assets-view.component.html',
  styleUrls: ['./assets-view.component.scss']
})
export class AssetsViewComponent implements OnInit, OnDestroy {
  assets: NetScanType[] = [];
  // defaultTime: ElasticFilterDefaultTime = new ElasticFilterDefaultTime('now-30d', 'now');
  pageWidth = window.innerWidth;
  filterWidth: number;
  tableWidth: number;
  sortEvent: any;
  totalItems: any;
  page = 0;
  loading = true;
  itemsPerPage = ITEMS_PER_PAGE;
  viewAssetDetail: NetScanType;
  sortBy = AssetFieldEnum.ASSET_ID + ',asc';
  assetsFields: UtmFieldType[] = ASSETS_FIELDS;
  checkbox: any;
  assetFieldEnum = AssetFieldEnum;
  fieldFilters = ASSETS_FIELDS_FILTERS;
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
  agentConsole: NetScanType;
  reasonRun: IncidentCommandType;
  agent: string;

  constructor(private utmNetScanService: UtmNetScanService,
              private modalService: NgbModal,
              private utmToastService: UtmToastService,
              private dataSourceInputService: DataSourceInputService,
              private router: Router,
              private spinner: NgxSpinnerService,
              private accountService: AccountService,
              private assetFiltersBehavior: AssetFiltersBehavior) {
  }

  // Init get asset on time filter component trigger


  ngOnInit() {
    this.setInitialWidth();
    this.getAssets();
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

  ngOnDestroy(): void {
    this.stopInterval(true);
    this.assetFiltersBehavior.$assetFilter.next(null);
  }

  setInitialWidth() {
    const dimensions = calcTableDimension(this.pageWidth);
    this.tableWidth = dimensions.tableWidth;
    this.filterWidth = dimensions.filterWidth;
  }

  loadPage($event: number) {
    this.requestParam.page = $event - 1;
    this.getAssets();
  }

  getAssets() {
    this.utmNetScanService.query(this.requestParam).subscribe(response => {
      this.totalItems = Number(response.headers.get('X-Total-Count'));
      this.assets = response.body;
      this.loading = false;
    });
  }

  onItemsPerPageChange($event: number) {
    this.itemsPerPage = $event;
    this.requestParam.size = $event;
    this.getAssets();
  }

  onTimeFilterChange($event: TimeFilterType) {
    this.requestParam.discoveredInitDate = $event.timeFrom;
    this.requestParam.discoveredEndDate = $event.timeTo;
    this.getAssets();
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
    this.getAssets();
  }

  toggleCheck() {
    this.checkbox = !this.checkbox;
    if (!this.checkbox) {
      this.assetsSelected = [];
    } else {
      this.assetsSelected = this.assets.map(value => value.id);
    }
  }

  addToSelected(event: Event, asset: NetScanType) {
    event.stopPropagation();
    const index = this.assetsSelected.findIndex(value => value === asset.id);
    if (index === -1) {
      this.assetsSelected.push(asset.id);
    } else {
      this.assetsSelected.splice(index, 1);
    }
  }

  isSelected(asset: NetScanType): boolean {
    return this.assetsSelected.findIndex(value => value === asset.id) !== -1;
  }

  onRowClicked(td: UtmFieldType, asset: NetScanType) {
    switch (td.field) {
      case AssetFieldEnum.ASSET_SEVERITY:
        break;
      case AssetFieldEnum.ASSET_METRICS:
        break;
      default:
        this.viewAssetDetail = asset;
    }
  }

  resetAllFilters() {
    for (const key of Object.keys(this.requestParam)) {
      if (!STATICS_FILTERS.includes(key)) {
        this.requestParam[key] = null;
      }
    }
    this.assetFiltersBehavior.$assetFilter.next(this.requestParam);
    this.getAssets();
  }

  onFilterChange($event: { prop: AssetFieldFilterEnum, values: any }) {
    switch ($event.prop) {
      case AssetFieldFilterEnum.PORTS:
        this.requestParam.openPorts = $event.values.length > 0 ? $event.values : null;
        break;
      case AssetFieldFilterEnum.SEVERITY:
        this.requestParam.severity = $event.values.length > 0 ? $event.values : null;
        break;
      case AssetFieldFilterEnum.TYPE:
        this.requestParam.type = $event.values.length > 0 ? $event.values : null;
        break;
      case AssetFieldFilterEnum.STATUS:
        this.requestParam.status = $event.values.length > 0 ? $event.values : null;
        break;
      case AssetFieldFilterEnum.ALIAS:
        this.requestParam.alias = $event.values.length > 0 ? $event.values : null;
        break;
      case AssetFieldFilterEnum.OS:
        this.requestParam.os = $event.values.length > 0 ? $event.values : null;
        break;
      case AssetFieldFilterEnum.ALIVE:
        this.requestParam.alive = $event.values;
        break;
      case AssetFieldFilterEnum.PROBE:
        this.requestParam.probe = $event.values.length > 0 ? $event.values : null;
        break;
      case AssetFieldFilterEnum.GROUP:
        this.requestParam.groups = $event.values.length > 0 ? $event.values : null;
        break;
    }
    this.assetFiltersBehavior.$assetAppliedFilter.next(this.requestParam);
    this.assetFiltersBehavior.$assetFilter.next(this.requestParam);
    this.getAssets();
  }

  onSearch($event: string) {
    this.requestParam.assetIpMacName = $event;
    this.requestParam.page = 0;
    this.getAssets();
  }

  deleteAsset(event: Event, asset: NetScanType) {
    event.stopPropagation();
    const deleteModalRef = this.modalService.open(ModalConfirmationComponent, {centered: true});
    deleteModalRef.componentInstance.header = 'Delete asset';
    deleteModalRef.componentInstance.message = 'Are you sure that you want to delete this source?';
    deleteModalRef.componentInstance.confirmBtnText = 'Delete';
    deleteModalRef.componentInstance.confirmBtnIcon = 'icon-display';
    deleteModalRef.componentInstance.confirmBtnType = 'delete';
    deleteModalRef.result.then(() => {
      this.delete(asset);
    });
  }

  delete(asset: NetScanType) {
    this.utmNetScanService.deleteCustomAsset(asset.id).subscribe(() => {
      this.utmToastService.showSuccessBottom('Asset deleted successfully');
      this.getAssets();
    }, () => {
      this.utmToastService.showError('Error deleting asset',
        'Error while trying to delete asset, please try again');
    });
  }


  deleteDataType(event: Event,dat: UtmDataInputStatus) {
    event.stopPropagation();
    this.deleting.push(dat.id);
    this.dataSourceInputService.delete(dat.id).subscribe(() => {
      this.getAssets();
      const indexDelete = this.deleting.indexOf(dat.id);
      if (indexDelete !== -1) {
        this.deleting.splice(indexDelete, 1);
      }
    });
  }


  getAssetSource(asset: NetScanType) {
    if (asset.assetName && asset.assetIp) {
      return asset.assetName + ' (' + asset.assetIp + ')';
    } else if (asset.assetName) {
      return asset.assetName;
    } else if (asset.assetIp) {
      return asset.assetIp;
    } else {
      return 'Unknown source';
    }
  }

  navigateToDataManagement(ip: string) {
    const queryParams = {alertType: 'ALERT'};
    queryParams[ALERT_SENSOR_FIELD] = ElasticOperatorsEnum.IS + ChartValueSeparator.BUCKET_SEPARATOR + ip;
    this.navigateWithParams('/data/alert/view', queryParams);
  }

  navigateWithParams(route: string, queryParams: object) {
    this.spinner.show('loadingSpinner');
    this.router.navigate([route], {
      queryParams
    }).then(() => {
      this.spinner.hide('loadingSpinner');
    });
  }


  getLastInput(asset: NetScanType) {
    if (asset.dataInputList.length > 0) {
      const lastInput = asset.dataInputList.sort((a, b) => a.timestamp > b.timestamp ? 1 : -1)[0].timestamp;
      return this.formatTimestampToDate(lastInput);
    } else {
      return 'Unknown';
    }
  }

  formatTimestampToDate(time: number) {
    const date = moment.unix(time);
    return moment.utc(date).format('YYYY-MM-DD HH:mm:ss');
  }


  toggleAsset(asset: NetScanType) {
    if (this.viewAssetDetail && this.viewAssetDetail.id === asset.id) {
      this.viewAssetDetail = undefined;
    } else {
      this.viewAssetDetail = asset;
    }
  }

  viwAgentDetail(event: Event, asset: NetScanType) {
    event.stopPropagation();
    this.viewAssetDetail = asset;
    this.agent = asset.assetName;
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
      this.getAssets();
    });
  }

  connectConsole(asset: NetScanType) {
    this.agentConsole = asset;
  }

  closeDetail() {
    this.agent = undefined;
    this.reasonRun.reason = '';
  }

  stopInterval(event: boolean) {
    if (event) {
      clearInterval(this.interval);
      this.interval = null;
    } else {
      this.starInterval();
    }
  }

  starInterval(){
    if (!this.interval) {
      this.interval = setInterval(() => {
        this.getAssets();
      }, 10000);
    }
  }
}
