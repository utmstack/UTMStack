import {HttpResponse} from '@angular/common/http';
import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {UtmModulesEnum} from '../../../../../app-module/shared/enum/utm-module.enum';
import {UtmModulesService} from '../../../../../app-module/shared/services/utm-modules.service';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
import {
  ALERT_CASE_ID_FIELD,
  ALERT_CATEGORY_FIELD,
  ALERT_FIELDS,
  ALERT_GENERATED_BY_FIELD,
  ALERT_ID_FIELD,
  ALERT_NAME_FIELD,
  ALERT_OBSERVATION_FIELD,
  ALERT_PROTOCOL_FIELD,
  ALERT_SENSOR_FIELD,
  ALERT_SEVERITY_FIELD_LABEL,
  ALERT_STATUS_FIELD,
  ALERT_TACTIC_FIELD,
  ALERT_TIMESTAMP_FIELD
} from '../../../../../shared/constants/alert/alert-field.constant';
import {LOG_ROUTE} from '../../../../../shared/constants/app-routes.constant';
import {LOG_INDEX_PATTERN, LOG_INDEX_PATTERN_ID} from '../../../../../shared/constants/main-index-pattern.constant';
import {ElasticOperatorsEnum} from '../../../../../shared/enums/elastic-operators.enum';
import {IncidentOriginTypeEnum} from '../../../../../shared/enums/incident-response/incident-origin-type.enum';
import {ElasticDataService} from '../../../../../shared/services/elasticsearch/elastic-data.service';
import {AlertTags} from '../../../../../shared/types/alert/alert-tag.type';
import {AlertStatusEnum, UtmAlertType} from '../../../../../shared/types/alert/utm-alert.type';
import {ElasticFilterType} from '../../../../../shared/types/filter/elastic-filter.type';
import {IncidentCommandType} from '../../../../../shared/types/incident/incident-command.type';
import {UtmFieldType} from '../../../../../shared/types/table/utm-field.type';
import {AlertUpdateHistoryBehavior} from '../../behavior/alert-update-history.behavior';
import {AlertUpdateTagBehavior} from '../../behavior/alert-update-tag.behavior';
import {AlertHistoryActionEnum} from '../../enums/alert-history-action.enum';
import {EventDataTypeEnum} from '../../enums/event-data-type.enum';
import {AlertManagementService} from '../../services/alert-management.service';

@Component({
  selector: 'app-alert-view-detail',
  templateUrl: './alert-view-detail.component.html',
  styleUrls: ['./alert-view-detail.component.scss']
})
export class AlertViewDetailComponent implements OnInit {
  @Input() alert: UtmAlertType;
  @Input() fullScreen = false;
  @Input() hideEmptyField = false;
  @Input() dataType: EventDataTypeEnum;
  @Input() tags: AlertTags[];
  @Output() refreshData = new EventEmitter<boolean>();
  ALERT_NAME = ALERT_NAME_FIELD;
  STATUS_FIELD = ALERT_STATUS_FIELD;
  SEVERITY_FIELD = ALERT_SEVERITY_FIELD_LABEL;
  OBSERVATION_FIELD = ALERT_OBSERVATION_FIELD;
  CATEGORY_FIELD = ALERT_CATEGORY_FIELD;
  SENSOR_FIELD = ALERT_SENSOR_FIELD;
  TIMESTAMP_FIELD = ALERT_TIMESTAMP_FIELD;
  PROTOCOL_FIELD = ALERT_PROTOCOL_FIELD;
  CASE_ID_FIELD = ALERT_CASE_ID_FIELD;
  GENERATED_BY_FIELD = ALERT_GENERATED_BY_FIELD;
  ALERT_TACTIC_FIELD = ALERT_TACTIC_FIELD;
  countRelatedEvents: number;
  viewHistory = false;
  viewLog = false;
  viewOnMap = true;
  alertActionEnum = AlertHistoryActionEnum;
  logs: string[];
  reference: string[] = [];
  log: any;
  getLastLog: boolean;
  viewRelatedRules = false;
  relatedTagsRules: number[];
  view = AlertDetailTabEnum.DETAIL;
  alertDetailTabEnum = AlertDetailTabEnum;
  socAi: boolean;
  commandJustification: IncidentCommandType;
  hideLastChange = false;
  incidentResponse = false;

  constructor(private alertServiceManagement: AlertManagementService,
              private utmToastService: UtmToastService,
              private alertUpdateHistoryBehavior: AlertUpdateHistoryBehavior,
              private alertUpdateTagBehavior: AlertUpdateTagBehavior,
              private spinner: NgxSpinnerService,
              private elasticDataService: ElasticDataService,
              private moduleService: UtmModulesService,
              private router: Router) {
  }

  ngOnInit() {
    this.viewLog = this.fullScreen;
    // @ts-ignore
    this.logs = this.alert.logs.reverse();
    this.countRelatedEvents = this.logs.length;
    const ref = this.alert.reference;
    this.reference = (ref && typeof ref !== 'string') ? ref : [];
    this.relatedTagsRules = this.alert.tagRulesApplied;
    this.moduleService.isActive(UtmModulesEnum.SOC_AI).subscribe(response => {
      this.socAi = response.body;
    });
    this.commandJustification = {
      originType: IncidentOriginTypeEnum.ALERT,
      originId: this.alert.id,
      reason: `Response to alert ${this.alert.name} with severity ${this.alert.severityLabel}`,
      command: ''
    };
  }

  getFieldByName(name): UtmFieldType {
    return ALERT_FIELDS.find(value => value.field === name);
  }

  isIgnoredAlert() {
    return AlertStatusEnum.IGNORED === this.alert.status;
  }

  isIncident(): boolean {
    return this.alert.isIncident;
  }

  onApplyTags($event: { tags: string[]; automatic: boolean }) {
    this.alertUpdateTagBehavior.$tagRefresh.next(true);
    const id = this.alert.id;
    this.alertUpdateTagBehavior.$updateTagForAlert.next(id);
    this.alertUpdateHistoryBehavior.$refreshHistory.next(true);
    this.refreshData.emit(true);
  }

  navigateToEvents() {
    const queryParams = {patternId: LOG_INDEX_PATTERN_ID, indexPattern: LOG_INDEX_PATTERN};
    const LOG_ID_FIELD = 'id';
    queryParams[LOG_ID_FIELD] = ElasticOperatorsEnum.IS_ONE_OF + '->' + this.logs.slice(0, 100);
    queryParams[this.TIMESTAMP_FIELD] = ElasticOperatorsEnum.IS_BETWEEN + '->' + 'now-1y' + ',' + 'now';
    this.spinner.show('loadingSpinner');
    this.router.navigate([LOG_ROUTE], {
      queryParams
    }).then(() => {
      this.spinner.hide('loadingSpinner');
    });
  }

  showMap(): boolean {
    const source = this.alert.source.coordinates;
    const destination = this.alert.destination.coordinates;
    return source && destination &&
      source.length > 0 && destination.length > 0
      && this.areCoordinatesNotZero(source) && this.areCoordinatesNotZero(destination);
  }

  areCoordinatesNotZero(coordinates: number[]): boolean {
    return coordinates[0] !== 0.0 || coordinates[1] !== 0.0;
  }


  searchLastLog() {
    if (!this.log) {
      this.getLastLog = true;
      const filter: ElasticFilterType[] = [{field: ALERT_ID_FIELD, operator: ElasticOperatorsEnum.IS, value: this.logs[0]}];
      this.elasticDataService.search(1, 1,
        1, LOG_INDEX_PATTERN, filter).subscribe(
        (res: HttpResponse<any>) => {
          this.log = res.body[0];
          this.getLastLog = false;
        },
        (res: HttpResponse<any>) => {
        }
      );
    }
  }

  viewLastLog() {
    this.searchLastLog();
  }
}

export enum AlertDetailTabEnum {
  DETAIL = 'detail',
  SOC_AI = 'socAi',
  HISTORY = 'history',
  LAST_LOG = 'lastLog',
  MAP = 'map',
  TAGS = 'tags',
  INCIDENT = 'incident',
}
