import {HttpResponse} from '@angular/common/http';
import { Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute, NavigationEnd, Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {TranslateService} from '@ngx-translate/core';
import {ResizeEvent} from 'angular-resizable-element';
import {NgxSpinnerService} from 'ngx-spinner';
import {Observable, Subject} from 'rxjs';
import {filter, takeUntil, tap} from 'rxjs/operators';
import {
  IrCreateRuleComponent
} from '../../../incident-response/shared/component/ir-create-rule/ir-create-rule.component';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {NewAlertBehavior} from '../../../shared/behaviors/new-alert.behavior';
import {
  ElasticFilterDefaultTime
} from '../../../shared/components/utm/filters/elastic-filter-time/elastic-filter-time.component';
import {
  ALERT_CASE_ID_FIELD,
  ALERT_FIELDS,
  ALERT_INCIDENT_FLAG_FIELD,
  ALERT_STATUS_FIELD,
  ALERT_STATUS_FIELD_AUTO,
  ALERT_STATUS_LABEL_FIELD,
  ALERT_TAGS_FIELD,
  ALERT_TIMESTAMP_FIELD,
  EVENT_FIELDS,
  EVENT_IS_ALERT,
  FALSE_POSITIVE_OBJECT,
  INCIDENT_FIELDS
} from '../../../shared/constants/alert/alert-field.constant';
import {AUTOMATIC_REVIEW, IGNORED} from '../../../shared/constants/alert/alert-status.constant';
import {ADMIN_ROLE} from '../../../shared/constants/global.constant';
import {MAIN_INDEX_PATTERN} from '../../../shared/constants/main-index-pattern.constant';
import {ITEMS_PER_PAGE} from '../../../shared/constants/pagination.constants';
import {SortDirection} from '../../../shared/directives/sortable/type/sort-direction.type';
import {SortEvent} from '../../../shared/directives/sortable/type/sort-event';
import {ElasticOperatorsEnum} from '../../../shared/enums/elastic-operators.enum';
import {DataNatureTypeEnum} from '../../../shared/enums/nature-data.enum';
import {ElasticDataService} from '../../../shared/services/elasticsearch/elastic-data.service';
import {CheckEmailConfigService, ParamShortType} from '../../../shared/services/util/check-email-config.service';
import {AlertTags} from '../../../shared/types/alert/alert-tag.type';
import {UtmAlertType} from '../../../shared/types/alert/utm-alert.type';
import {ElasticFilterType} from '../../../shared/types/filter/elastic-filter.type';
import {UtmFieldType} from '../../../shared/types/table/utm-field.type';
import {TimeFilterType} from '../../../shared/types/time-filter.type';
import {invertFilter, mergeParams, sanitizeFilters} from '../../../shared/util/elastic-filter.util';
import {parseQueryParamsToFilter} from '../../../shared/util/query-params-to-filter.util';
import {SaveAlertReportComponent} from '../alert-reports/shared/components/save-report/save-report.component';
import {AlertDataTypeBehavior} from '../shared/behavior/alert-data-type.behavior';
import {AlertFiltersBehavior} from '../shared/behavior/alert-filters.behavior';
import {AlertStatusBehavior} from '../shared/behavior/alert-status.behavior';
import {RowToFiltersComponent} from '../shared/components/filters/row-to-filter/row-to-filters.component';
import {EventDataTypeEnum} from '../shared/enums/event-data-type.enum';
import {AlertTagService} from '../shared/services/alert-tag.service';
import {OpenAlertsService} from '../shared/services/open-alerts.service';
import {getCurrentAlertStatus, getStatusName} from '../shared/util/alert-util-function';


@Component({
  selector: 'app-alert-view',
  templateUrl: './alert-view.component.html',
  styleUrls: ['./alert-view.component.scss']
})
export class AlertViewComponent implements OnInit, OnDestroy {
  fields = ALERT_FIELDS;
  manageTags: number;
  ADMIN = ADMIN_ROLE;
  alerts: UtmAlertType[] = [];
  tableWidth: number;
  checkbox = false;
  loading = true;
  /**
   * Contains ID of selected alerts
   */
  alertSelected: UtmAlertType[] = [];
  statusChange: number;
  //
  direction: SortDirection = '';
  sortBy = ALERT_TIMESTAMP_FIELD + ',desc';
  alertDetail: UtmAlertType;
  viewAlertDetail: boolean;
  totalItems: any;
  page = 1;
  itemsPerPage = ITEMS_PER_PAGE;
  // By default all alert will contain all except alerts in review
  filters: ElasticFilterType[] = [
    {field: ALERT_STATUS_FIELD_AUTO, operator: ElasticOperatorsEnum.IS_NOT, value: AUTOMATIC_REVIEW},
    {field: ALERT_TAGS_FIELD, operator: ElasticOperatorsEnum.IS_NOT, value: FALSE_POSITIVE_OBJECT.tagName},
    {field: ALERT_TIMESTAMP_FIELD, operator: ElasticOperatorsEnum.IS_BETWEEN, value: ['now-7d', 'now']}
  ];
  defaultStatus: number;
  dataNature = DataNatureTypeEnum.ALERT;
  sortEvent: SortEvent;
  pattern = MAIN_INDEX_PATTERN;
  //
  defaultTime: ElasticFilterDefaultTime;
  IGNORED = IGNORED;
  pageWidth = window.innerWidth;
  filterWidth: number;
  incomingAlert$: Observable<number>;
  dataType: EventDataTypeEnum;
  eventDataTypeEnum = EventDataTypeEnum;
  refreshingAlert = false;
  firstLoad = true;
  tags: AlertTags[];
  showRefresh = false;
  destroy$ = new Subject<void>();

  constructor(private elasticDataService: ElasticDataService,
              private modalService: NgbModal,
              private utmToastService: UtmToastService,
              private translate: TranslateService,
              private alertFiltersBehavior: AlertFiltersBehavior,
              private updateStatusServiceBehavior: AlertStatusBehavior,
              private activatedRoute: ActivatedRoute,
              public router: Router,
              private openAlertsService: OpenAlertsService,
              private alertDataTypeBehavior: AlertDataTypeBehavior,
              private alertTagService: AlertTagService,
              private spinner: NgxSpinnerService,
              private checkEmailConfigService: CheckEmailConfigService) {
    // this.tableWidth = this.pageWidth - 300;
  }

  ngOnInit() {
    this.checkEmailConfigService.check(ParamShortType.Alert);
    this.setInitialWidth();
    this.getTags();
    this.activatedRoute.queryParams.subscribe(params => {
      const queryParams = Object.entries(params).length > 0 ? params : null;
      if (queryParams) {
        parseQueryParamsToFilter(queryParams).then((filter) => {
          mergeParams(filter, this.filters).then((filters) => {
            this.filters = filters;
            if (queryParams.alertType) {
              this.dataType = queryParams.alertType;
              this.fields = this.resolveFieldsByDataType(this.dataType);
              this.alertDataTypeBehavior.$alertDataType.next(this.dataType);
              this.changeParamsByDataType(this.dataType).then(() => this.setDefaultParams());
            } else {
              this.resolveDataType().then(type => {
                this.dataType = type;
                this.fields = this.resolveFieldsByDataType(this.dataType);
                this.alertDataTypeBehavior.$alertDataType.next(this.dataType);
                /**
                 * If type is event then dont show status field
                 */
                this.changeParamsByDataType(type).then(() => this.setDefaultParams());
              });
            }
            this.setDefaultParams();
          });
        });
      } else {
        /**
         * If does not exist any param filter, set default to EVENT
         */
        this.dataType = EventDataTypeEnum.EVENT;
        this.fields = EVENT_FIELDS;
        this.alertDataTypeBehavior.$alertDataType.next(this.dataType);
        /**
         * Emit default array filter
         */
        this.alertFiltersBehavior.$filters.next(this.filters);
      }
    });
    this.incomingAlert$ = this.openAlertsService.openAlerts$
      .pipe(
        takeUntil(this.destroy$),
        tap(() => this.showRefresh = true));
  }

  refreshAlerts() {
    this.refreshingAlert = true;
    this.page = 1;
    this.getAlert();
    this.firstLoad = false;
    this.showRefresh = false;
  }

  getTags() {
    this.alertTagService.query({page: 0, size: 1000}).subscribe(reponse => {
      this.tags = reponse.body;
    });
  }

  resolveFieldsByDataType(dataType: EventDataTypeEnum): UtmFieldType[] {
    switch (dataType) {
      case EventDataTypeEnum.INCIDENT:
        return INCIDENT_FIELDS;
      case EventDataTypeEnum.EVENT:
        return EVENT_FIELDS;
      case EventDataTypeEnum.ALERT:
        return ALERT_FIELDS;
      case EventDataTypeEnum.FALSE_POSITIVE:
        return ALERT_FIELDS;
    }
  }

  changeParamsByDataType(dataType: EventDataTypeEnum): Promise<boolean> {
    return new Promise<boolean>(resolve => {
      if (dataType === EventDataTypeEnum.EVENT) {
        this.fields = this.fields.filter(value => value.field !== ALERT_STATUS_LABEL_FIELD);
        this.clearFilter(ALERT_INCIDENT_FLAG_FIELD).then(() => resolve(true));
        this.clearFilter(EVENT_IS_ALERT).then(() => resolve(true));
      } else if (dataType === EventDataTypeEnum.ALERT) {
        this.clearFilter(ALERT_INCIDENT_FLAG_FIELD).then(() => resolve(true));
      } else if (dataType === EventDataTypeEnum.INCIDENT) {
        this.clearFilter(EVENT_IS_ALERT).then(() => resolve(true));
      } else if (dataType === EventDataTypeEnum.FALSE_POSITIVE) {
        this.clearFilter(ALERT_INCIDENT_FLAG_FIELD).then(() => resolve(true));
        const indexFilter = this.filters.findIndex(value => value.field === ALERT_TAGS_FIELD);
        if (indexFilter !== -1) {
          this.filters[indexFilter] = invertFilter(this.filters[indexFilter]);
        }
      }
    });
  }


  clearFilter(field): Promise<boolean> {
    return new Promise<boolean>(resolve => {
      const deleteIndexFilter = this.filters.findIndex(value => value.field.includes(field));
      if (deleteIndexFilter !== -1) {
        this.filters.splice(deleteIndexFilter, 1);
      }
      resolve(true);
    });
  }

  /**
   * Return EventDataTypeEnum based on filters
   * Check if exist any incident field with operator IS, if exist return INCIDENT
   * Check if exist any isAlert field with operator IS, if exist return ALERT
   * else return EVENT
   */
  resolveDataType(): Promise<EventDataTypeEnum> {
    return new Promise<EventDataTypeEnum>(resolve => {
      const indexOfIsIncident = this.filters.findIndex(value => value.field === ALERT_INCIDENT_FLAG_FIELD
        && value.operator === ElasticOperatorsEnum.IS);
      const indexOfIsAlert = this.filters.findIndex(value => value.field === EVENT_IS_ALERT
        && value.operator === ElasticOperatorsEnum.IS);
      if (indexOfIsIncident !== -1) {
        resolve(EventDataTypeEnum.INCIDENT);
      } else if (indexOfIsAlert !== -1) {
        resolve(EventDataTypeEnum.ALERT);
      } else {
        resolve(EventDataTypeEnum.EVENT);
      }
    });
  }

  /**
   * Resolve field based on EventDataTypeEnum
   * @param type EventDataTypeEnum
   */
  resolveFieldByDataTypeEnum(type: EventDataTypeEnum) {
    switch (type) {
      case EventDataTypeEnum.ALERT:
        return EVENT_IS_ALERT;
      case EventDataTypeEnum.EVENT:
        return null;
      case EventDataTypeEnum.INCIDENT:
        return ALERT_INCIDENT_FLAG_FIELD;
    }

  }

  /**
   * After merge filter set default values to params new values
   */
  setDefaultParams() {
    const indexTime = this.filters.findIndex(value => value.field === ALERT_TIMESTAMP_FIELD);
    if (indexTime !== -1) {
      this.defaultTime = new ElasticFilterDefaultTime(this.filters[indexTime].value[0], this.filters[indexTime].value[1]);
    } else {
      this.defaultTime = new ElasticFilterDefaultTime('now-7d', 'now');
    }
    this.getCurrentStatus();
    this.getAlert('on set default params');
    this.updateStatusServiceBehavior.$updateStatus.next(true);
    this.alertFiltersBehavior.$filters.next(this.filters);
  }

  /**
   * Return current status value from filter
   */
  getCurrentStatus(): number {
    return getCurrentAlertStatus(this.filters);
  }

  onTimeFilterChange($event: TimeFilterType) {
    const timeFilterIndex = this.filters.findIndex(value => value.field === ALERT_TIMESTAMP_FIELD);
    if (timeFilterIndex === -1) {
      this.filters.push({
        field: ALERT_TIMESTAMP_FIELD,
        value: [$event.timeFrom, $event.timeTo],
        operator: ElasticOperatorsEnum.IS_BETWEEN
      });
    } else {
      this.filters[timeFilterIndex].value = [$event.timeFrom, $event.timeTo];
    }
    this.alertFiltersBehavior.$filters.next(this.filters);
    this.getAlert('on time filter change');
  }

  getAlert(calledFrom?: string) {
    this.elasticDataService.search(this.page, this.itemsPerPage,
      100000000, this.dataNature,
      sanitizeFilters(this.filters), this.sortBy).subscribe(
      (res: HttpResponse<any>) => {
        this.totalItems = Number(res.headers.get('X-Total-Count'));
        this.alerts = res.body;
        this.loading = false;
        this.refreshingAlert = false;
      },
      (res: HttpResponse<any>) => {
        this.utmToastService.showError('Error', 'An error occurred while listing the alerts. Please try again later.');
      }
    );
  }

  saveReport() {
    const modalSaveReport = this.modalService.open(SaveAlertReportComponent,
      {centered: true});
    modalSaveReport.componentInstance.filters = this.filters;
    modalSaveReport.componentInstance.dataType = this.dataType;
    modalSaveReport.componentInstance.fields = this.fields.filter(value => value.visible === true);
  }

  toggleCheck() {
    this.checkbox = !this.checkbox;
    if (!this.checkbox) {
      this.alertSelected = [];
    } else {
      for (const alert of this.alerts) {
        const index = this.alertSelected.indexOf(alert);
        if (index === -1) {
          this.alertSelected.push(alert);
        }
      }
    }
  }

  addToSelected(alert: any) {
    const index = this.alertSelected.indexOf(alert);
    if (index === -1) {
      this.alertSelected.push(alert);
    } else {
      this.alertSelected.splice(index, 1);
    }
  }

  isSelected(alert: any): boolean {
    return this.alertSelected.findIndex(value => value.id === alert.id) !== -1;
  }

  onSortBy($event: SortEvent) {
    this.sortBy = $event.column + ',' + $event.direction;
    this.getAlert('on sort by');
  }

  getRowToFiltersData(alert: any) {
    const modalRef = this.modalService.open(RowToFiltersComponent, {centered: true});
    modalRef.componentInstance.alert = alert;
    modalRef.componentInstance.fields = this.fields;
    modalRef.componentInstance.addRowToFilter.subscribe(filterRow => {
      mergeParams(filterRow, this.filters).then(value => {
        this.filters = value;
        this.page = 1;
        this.getAlert('on add row to filter');
        this.alertFiltersBehavior.$filters.next(this.filters);
      });
    });
  }


  loadPage(page: any) {
    this.page = page;
    this.getAlert('on load page');
  }

  onFilterStatusChange($event: ElasticFilterType) {
    const indexStatus = this.filters.findIndex(value => value.field === $event.field);
    if (indexStatus === -1) {
      this.filters.push($event);
    } else {
      this.filters[indexStatus] = $event;
      for (let i = 0; i < this.filters.length; i++) {
        if (this.filters[i].field.includes(ALERT_STATUS_LABEL_FIELD)) {
          this.filters.splice(i, 1);
        }
      }
    }
    this.alertFiltersBehavior.$filters.next(this.filters);
    this.page = 1;
    this.getAlert('on status filter change');
  }

  onApplyStatusChange($event: number) {
    this.showStatusChangeToast($event);
  }

  showStatusChangeToast(status) {
    const msg = getStatusName(status);
    this.getAlert('on change status');
    this.updateStatusServiceBehavior.$updateStatus.next(true);
    this.translate.get(['toast.changeAlertStatus', msg]).subscribe(value => {
      this.utmToastService.showSuccessBottom(value['toast.changeAlertStatus'] + ' ' + value[msg].toString().toUpperCase());
    });
  }

  onApplyTag($event: { tags: string[], automatic: boolean }) {
    this.getAlert('on apply tag');
  }

  onItemsPerPageChange($event: number) {
    this.itemsPerPage = $event;
    this.page = 1;
    this.getAlert('on change items per page');
  }

  onFilterChange($event: ElasticFilterType) {
    this.processFilters($event).then(filters => {
      this.filters = filters;
      this.page = 1;
      this.getAlert('on generic filter change');
      this.updateStatusServiceBehavior.$updateStatus.next(true);
      this.alertFiltersBehavior.$filters.next(this.filters);
    });
  }

  processFilters(filter: ElasticFilterType): Promise<ElasticFilterType[]> {
    return new Promise<ElasticFilterType[]>(resolve => {
      const indexFilters = this.filters.findIndex(value => filter.field.includes(value.field));
      if (indexFilters === -1) {
        this.filters.push(filter);
      } else {
        this.filters[indexFilters] = filter;
      }
      resolve(this.filters);
    });
  }

  getRuleName(): string {
    return this.alertDetail.name;
  }

  viewDetailAlert(alert: any, td: UtmFieldType) {
    if (td.field !== ALERT_STATUS_FIELD) {
      this.alertDetail = alert;
      this.viewAlertDetail = true;
    }
  }

  onRefreshData($event: boolean) {
    this.getAlert('on refresh data');
    this.updateStatusServiceBehavior.$updateStatus.next(true);
    this.getTags();
  }

  onApplyNote(alert: any, $event: string) {
    // this.alertServiceManagement.updateNotes(getID(alert), $event).subscribe(response => {
    //   this.utmToastService.showSuccessBottom('Note added successfully');
    //   this.getAlert();
    // });
  }

  setInitialWidth() {
    if (this.pageWidth > 1980) {
      this.filterWidth = 350;
      this.tableWidth = this.pageWidth - this.filterWidth - 51;
    } else {
      this.filterWidth = 300;
      this.tableWidth = this.pageWidth - this.filterWidth - 51;
    }
    if (this.pageWidth > 2500) {
      this.filterWidth = 350;
      this.tableWidth = this.pageWidth - this.filterWidth - 51;
    }
    if (this.pageWidth > 4000) {
      this.filterWidth = 400;
      this.tableWidth = this.pageWidth - this.filterWidth - 51;
    }
  }

  onFilterAppliedChange($event: { filter: ElasticFilterType, valueDelete: string }) {
    this.processFilters($event.filter).then(filters => {
      this.alertFiltersBehavior.$deleteFilterValue.next({field: $event.filter.field, value: $event.valueDelete});
      this.filters = filters;
      this.updateStatusServiceBehavior.$updateStatus.next(true);
      this.page = 1;
      this.getAlert('on filter applied change');
    });
  }

  onResize($event: ResizeEvent) {
    if ($event.rectangle.width >= 250) {
      this.tableWidth = (this.pageWidth - $event.rectangle.width - 51);
      this.filterWidth = $event.rectangle.width;
    }
  }

  navigateTo(link: string) {
    this.spinner.show('loadingSpinner');
    this.router.navigate([link]).then(() => {
      this.spinner.hide('loadingSpinner');
    });
  }

  loadNewAlerts() {
    this.page = 1;
    this.sortBy = ALERT_CASE_ID_FIELD + ',desc';
    this.getAlert('on load new alerts');
  }

  onFilterReset($event: boolean) {
    const queryParams = {};
    if (this.dataType !== this.eventDataTypeEnum.EVENT) {
      const field = 'alertType';
      queryParams[field] = 'ALERT';
    }
    this.router.navigate(['/data/alert/view'], {
      queryParams
    }).then(() => {
      /**
       * Filter reset consist in filter current array filter by datatype and time, so
       * will exist only two filters
       */
      this.filters = this.filters.filter(value => value.field === this.resolveFieldByDataTypeEnum(this.dataType) ||
        value.field === ALERT_TIMESTAMP_FIELD);
      this.setDefaultParams();
      this.alertFiltersBehavior.$resetFilter.next(true);
    });
  }

  onSuccessMarkAsIncident($event: string) {
    this.alertSelected = [];
    this.getAlert('on marked as incident');
  }

  getAlertsIds() {
    return this.alertSelected.map(value => value.id);
  }

  openIncidentResponseAutomationModal(alert: UtmAlertType) {
    const modal = this.modalService.open(IrCreateRuleComponent, {size: 'lg', centered: true});
    modal.componentInstance.alert = alert;
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
