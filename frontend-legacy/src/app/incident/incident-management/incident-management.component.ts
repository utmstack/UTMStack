import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {ResizeEvent} from 'angular-resizable-element';
import {STATICS_FILTERS} from '../../assets-discover/shared/const/filter-const';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {ITEMS_PER_PAGE} from '../../shared/constants/pagination.constants';
import {SortEvent} from '../../shared/directives/sortable/type/sort-event';
import {IncidentOriginTypeEnum} from '../../shared/enums/incident-response/incident-origin-type.enum';
import {IncidentSeverityEnum} from '../../shared/enums/incident/incident-severity.enum';
import {IncidentStatusEnum} from '../../shared/enums/incident/incident-status.enum';
import {UtmIncidentService} from '../../shared/services/incidents/utm-incident.service';
import {IncidentCommandType} from '../../shared/types/incident/incident-command.type';
import {IncidentFilterType} from '../../shared/types/incident/incident-filter.type';
import {UtmIncidentType} from '../../shared/types/incident/utm-incident.type';
import {TimeFilterType} from '../../shared/types/time-filter.type';
import {calcTableDimension} from '../../shared/util/screen.util';
import {CheckEmailConfigService, ParamShortType} from "../../shared/services/util/check-email-config.service";

@Component({
  selector: 'app-incident-management',
  templateUrl: './incident-management.component.html',
  styleUrls: ['./incident-management.component.scss']
})
export class IncidentManagementComponent implements OnInit, OnDestroy {
  incidents: UtmIncidentType[];
  incidentDetail: UtmIncidentType;
  pageWidth = window.innerWidth;
  filterWidth: number;
  tableWidth: number;
  sortEvent: any;
  totalItems: any;
  page = 0;
  loading = true;
  itemsPerPage = ITEMS_PER_PAGE;
  viewIncidentHistory: UtmIncidentType;
  viewIncidentNotes: UtmIncidentType;
  executeCommandIncident: UtmIncidentType;
  sortBy = 'id,desc';
  requestParam: IncidentFilterType = {
    page: 0,
    size: ITEMS_PER_PAGE,
    sort: 'id,desc',
  };
  interval: any;
  checkbox: any;
  display = 0;
  severities: { label: string, value: number }[] = [
    {label: 'Low', value: IncidentSeverityEnum.LOW},
    {label: 'Medium', value: IncidentSeverityEnum.MEDIUM},
    {label: 'High', value: IncidentSeverityEnum.HIGH},
  ];
  status = [IncidentStatusEnum.COMPLETED, IncidentStatusEnum.IN_REVIEW, IncidentStatusEnum.OPEN];
  usersAssigned: { id: number, login: string }[];
  loadingUsers = true;
  incidentId = 0;
  incidentTabEnum = IncidentTabEnum;
  tabSelected: IncidentTabEnum = IncidentTabEnum.DETAIL;
  reasonRun: IncidentCommandType;

  constructor(private utmIncidentService: UtmIncidentService,
              private modalService: NgbModal,
              private activatedRoute: ActivatedRoute,
              private checkEmailConfigService: CheckEmailConfigService) {

  }

  ngOnInit() {
    this.checkEmailConfigService.check(ParamShortType.Incident);
    this.setInitialWidth();
    this.getIncidents();
    this.activatedRoute.queryParams.subscribe(params => {
      if (params.incidentId) {
        this.requestParam['id.equals'] = params.incidentId;
        this.incidentId = Number(params.incidentId);
        this.getIncidents();
      }
    });
    this.interval = setInterval(() => {
      this.getIncidents();
    }, 60000);

  }

  ngOnDestroy(): void {
    clearInterval(this.interval);
  }

  setInitialWidth() {
    const dimensions = calcTableDimension(this.pageWidth);
    this.filterWidth = dimensions.filterWidth;
    this.tableWidth = dimensions.tableWidth;
  }

  loadPage($event: number) {
    this.requestParam.page = $event - 1;
    this.getIncidents();
  }

  getUsersAssigned() {
    this.utmIncidentService.getUsersAssigned().subscribe(response => {
      this.usersAssigned = response.body;
      this.loadingUsers = false;
    });
  }

  getIncidents() {
    this.utmIncidentService.query(this.requestParam).subscribe(response => {
      this.totalItems = Number(response.headers.get('X-Total-Count'));
      this.incidents = response.body;
      this.loading = false;
    });
  }

  onItemsPerPageChange($event: number) {
    this.itemsPerPage = $event;
    this.requestParam.size = $event;
    this.getIncidents();
  }

  onTimeFilterChange($event: TimeFilterType) {
    this.requestParam['incidentCreatedDate.greaterThanOrEqual'] = $event.timeFrom;
    this.requestParam['incidentCreatedDate.lessThanOrEqual'] = $event.timeTo;
    this.getIncidents();
  }

  onResize($event: ResizeEvent) {
    if ($event.rectangle.width >= 250) {
      this.tableWidth = (this.pageWidth - $event.rectangle.width - 51);
      this.filterWidth = $event.rectangle.width;
    }
  }

  resetAllFilters() {
    for (const key of Object.keys(this.requestParam)) {
      if (!STATICS_FILTERS.includes(key)) {
        this.requestParam[key] = null;
      }
    }
    this.getIncidents();
  }

  onSearch($event: string) {
    this.requestParam['incidentName.contains'] = $event;
    this.requestParam.page = 0;
    this.getIncidents();
  }

  toggleCheck() {

  }

  isSelected(incident: UtmIncidentType) {
    return true;
  }

  onSortBy($event: SortEvent) {
    this.requestParam.sort = $event.column + ',' + $event.direction;
    this.getIncidents();
  }

  viewHistory(incident: UtmIncidentType) {
    this.viewIncidentHistory = incident;
  }

  onSearchByID($event: number) {
    if ($event) {
      this.requestParam['id.equals'] = Number($event);
      this.requestParam.page = 0;
      this.getIncidents();
    } else {
      this.requestParam['id.equals'] = null;
      this.requestParam.page = 0;
      this.getIncidents();
    }
  }

  filterBySeverity(value: number) {
    const sev = this.requestParam['incidentSeverity.in'] ? this.requestParam['incidentSeverity.in'] : [];
    const index = sev.indexOf(value);
    if (index !== -1) {
      sev.splice(index, 1);
    } else {
      sev.push(value);
    }
    this.requestParam['incidentSeverity.in'] = sev.length ? sev : null;
    this.requestParam.page = 0;
    this.getIncidents();
  }

  filterByStatus(stat: IncidentStatusEnum) {
    const status = this.requestParam['incidentStatus.in'] ? this.requestParam['incidentStatus.in'] : [];
    const index = status.indexOf(stat);
    if (index !== -1) {
      status.splice(index, 1);
    } else {
      status.push(stat);
    }
    this.requestParam['incidentStatus.in'] = status.length ? status : null;
    this.requestParam.page = 0;
    this.getIncidents();
  }

  filterByUserAssigned(user: { id: number; login: string }) {
  }

  showAlertForIncident(incident: UtmIncidentType) {
    this.incidentDetail = incident;
    this.reasonRun = {
      command: '',
      reason: 'Response to incident' + this.incidentDetail.incidentName +
        ' with severity ' + this.getIncidentSeverityLabel(this.incidentDetail.incidentSeverity),
      originId: this.incidentDetail.id.toString(),
      originType: IncidentOriginTypeEnum.INCIDENT
    };
  }

  createHistory() {

  }

  onStatusChange($event: IncidentStatusEnum) {
  }

  private getIncidentSeverityLabel(severity: number): string {
    switch (severity) {
      case IncidentSeverityEnum.HIGH:
        return 'HIGH';
      case IncidentSeverityEnum.MEDIUM:
        return 'MEDIUM';
      case IncidentSeverityEnum.LOW:
        return 'LOW';
      default:
        return 'UNKNOWN';
    }
  }

}


export enum IncidentTabEnum {
  DETAIL = 'detail',
  HISTORY = 'history',
  NOTES = 'notes',
  COMMAND = 'command',
}
