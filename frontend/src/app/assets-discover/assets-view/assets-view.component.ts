import {HttpResponse} from '@angular/common/http';
import {ChangeDetectionStrategy, Component, OnDestroy, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {ResizeEvent} from 'angular-resizable-element';
import * as moment from 'moment';
import {NgxSpinnerService} from 'ngx-spinner';
import {Observable} from 'rxjs';
import {filter, map, switchMap, tap} from 'rxjs/operators';
import {AccountService} from '../../core/auth/account.service';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {
  ModalConfirmationComponent
} from '../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {ALERT_SENSOR_FIELD} from '../../shared/constants/alert/alert-field.constant';
import {ITEMS_PER_PAGE} from '../../shared/constants/pagination.constants';
import {SortEvent} from '../../shared/directives/sortable/type/sort-event';
import {ChartValueSeparator} from '../../shared/enums/chart-value-separator';
import {ElasticOperatorsEnum} from '../../shared/enums/elastic-operators.enum';
import {IncidentOriginTypeEnum} from '../../shared/enums/incident-response/incident-origin-type.enum';
import {UtmDatePipe} from '../../shared/pipes/date.pipe';
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
import {SortDirection} from "../../shared/directives/sortable/type/sort-direction.type";

@Component({
  selector: 'app-assets-view',
  templateUrl: './assets-view.component.html',
  styleUrls: ['./assets-view.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class AssetsViewComponent implements OnInit, OnDestroy {
  assets$: Observable<NetScanType[]>;
  assets: NetScanType[];
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
    groups: null,
    dataTypes: null,
  };
  assetsSelected: number[] = [];
  interval: any;
  deleting: string[] = [];
  agentConsole: NetScanType;
  reasonRun: IncidentCommandType;
  agent: string;
  noData = false;

  constructor(private utmNetScanService: UtmNetScanService,
              private modalService: NgbModal,
              private utmToastService: UtmToastService,
              private dataSourceInputService: DataSourceInputService,
              private router: Router,
              private spinner: NgxSpinnerService,
              private accountService: AccountService,
              private assetFiltersBehavior: AssetFiltersBehavior,
              private datePipe: UtmDatePipe) {
  }

  ngOnInit() {
    this.setInitialWidth();

    this.accountService.identity().then(account => {
      this.reasonRun = {
        command: '',
        reason: '',
        originId: account.login,
        originType: IncidentOriginTypeEnum.USER_EXECUTION
      };
    });

    this.assets$ = this.utmNetScanService.onRefresh$
      .pipe(
        filter(refresh => !!refresh),
        switchMap(() => this.utmNetScanService.fetchData(this.requestParam)),
        tap((response: HttpResponse<NetScanType[]>) => {
          this.totalItems = Number(response.headers.get('X-Total-Count'));
          this.loading = false;
          this.assets = response.body;
          this.noData = response.body.length === 0;
        }),
        map((response) => {
          return response.body.map(asset => {
            if (asset.dataInputList && asset.dataInputList.length > 0) {
              asset.dataInputList = asset.dataInputList.sort((a, b) => a.timestamp - b.timestamp);
            } else {
              asset.dataInputList = [];
            }

            const displayName = asset.assetName && asset.assetIp ? `${asset.assetName} (${asset.assetIp})`
              : asset.assetName ? asset.assetName : asset.assetIp ? asset.assetIp : 'Unknown source';

            const sortKey = (asset.assetName || '') + (asset.assetIp || '');

            return { ...asset, displayName, sortKey };
          });
        })
      );

    this.utmNetScanService.notifyRefresh(true);
    //this.starInterval();
  }

  setInitialWidth() {
    const dimensions = calcTableDimension(this.pageWidth);
    this.tableWidth = dimensions.tableWidth;
    this.filterWidth = dimensions.filterWidth;
  }

  loadPage(page: number) {
    this.page = page - 1;
    this.requestParam.page = this.page;
    this.utmNetScanService.notifyRefresh(true);
  }

  onItemsPerPageChange($event: number) {
    this.itemsPerPage = $event;
    this.requestParam.size = $event;
    this.utmNetScanService.notifyRefresh(true);
  }

  onTimeFilterChange($event: TimeFilterType) {
    this.requestParam.discoveredInitDate = $event.timeFrom;
    this.requestParam.discoveredEndDate = $event.timeTo;
    this.utmNetScanService.notifyRefresh(true);
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
    if ($event.column === 'displayName') {
      this.sortAssets($event.direction);
    } else {
      this.requestParam.sort = $event.column + ',' + $event.direction;
      this.getAssets();
    }
  }

  sortAssets(direction: SortDirection) {
    this.assets.sort((a, b) => {

      if (a.displayName === 'Unknown source') { return 1; }
      if (b.displayName === 'Unknown source') { return -1; }

      const aVal = a.sortKey;
      const bVal = b.sortKey;

      const ipRegex = /^(\d{1,3}\.){3}\d{1,3}$/;
      const aIsIP = ipRegex.test(aVal);
      const bIsIP = ipRegex.test(bVal);

      if (aIsIP && bIsIP) {
        const aOctets = aVal.split('.').map(Number);
        const bOctets = bVal.split('.').map(Number);
        for (let i = 0; i < 4; i++) {
          if (aOctets[i] !== bOctets[i]) { return direction === 'asc' ? aOctets[i] - bOctets[i] : bOctets[i] - aOctets[i]; }
        }
        return 0;
      }

      if (aIsIP) { return direction === 'asc' ? -1 : 1; }
      if (bIsIP) { return direction === 'asc' ? 1 : -1; }

      const cmp = aVal.localeCompare(bVal, undefined, { numeric: true, sensitivity: 'base' });
      return direction === 'asc' ? cmp : -cmp;
    });
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
    this.utmNetScanService.notifyRefresh(true);
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
        this.requestParam.alive = $event.values.length > 0 ? $event.values : null;
        break;
      case AssetFieldFilterEnum.PROBE:
        this.requestParam.probe = $event.values.length > 0 ? $event.values : null;
        break;
      case AssetFieldFilterEnum.GROUP:
        this.requestParam.groups = $event.values.length > 0 ? $event.values : null;
        break;
      case AssetFieldFilterEnum.DATA_TYPES:
        this.requestParam.dataTypes = $event.values.length > 0 ? $event.values : null;
        break;
    }
    this.assetFiltersBehavior.$assetAppliedFilter.next(this.requestParam);
    this.assetFiltersBehavior.$assetFilter.next(this.requestParam);
    this.utmNetScanService.notifyRefresh(true);
  }

  onSearch($event: string) {
    this.requestParam.assetIpMacName = $event;
    this.requestParam.page = 0;
    this.utmNetScanService.notifyRefresh(true);
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
      this.utmNetScanService.notifyRefresh(true);
    }, () => {
      this.utmToastService.showError('Error deleting asset',
        'Error while trying to delete asset, please try again');
    });
  }


  deleteDataType(event: Event, dat: UtmDataInputStatus) {
    event.stopPropagation();
    this.deleting.push(dat.id);
    this.dataSourceInputService.delete(dat.id).subscribe(() => {
      this.utmNetScanService.notifyRefresh(true);
      const indexDelete = this.deleting.indexOf(dat.id);
      if (indexDelete !== -1) {
        this.deleting.splice(indexDelete, 1);
      }
    });
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
      return this.datePipe.transform(this.formatTimestampToDate(lastInput));
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
      this.utmNetScanService.notifyRefresh(true);
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

  starInterval() {
    if (!this.interval) {
      this.interval = setInterval(() => {
        this.utmNetScanService.notifyRefresh(true);
      }, 60000);
    }
  }

  getAssets() {
    this.utmNetScanService.notifyRefresh(true);
  }

  trackByFn(index: number, item: any) {
    return item.id;
  }

  trackByDataInputFn(index: number, item: UtmDataInputStatus) {
    return item.id;
  }

  ngOnDestroy(): void {
    this.stopInterval(true);
    this.assetFiltersBehavior.$assetFilter.next(null);
  }
}
