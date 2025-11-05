import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {UtmModulesEnum} from '../../../../../app-module/shared/enum/utm-module.enum';
import {UtmModulesService} from '../../../../../app-module/shared/services/utm-modules.service';
import {
  UtmTableDetailComponent
} from '../../../../../shared/components/utm/table/utm-table/utm-table-detail/utm-table-detail.component';
import {
  ALERT_ADVERSARY_FIELD,
  ALERT_CASE_ID_FIELD,
  ALERT_CATEGORY_FIELD,
  ALERT_FIELDS,
  ALERT_GENERATED_BY_FIELD, ALERT_IMPACT_FIELD,
  ALERT_NAME_FIELD,
  ALERT_OBSERVATION_FIELD,
  ALERT_PROTOCOL_FIELD,
  ALERT_SENSOR_FIELD,
  ALERT_SEVERITY_FIELD_LABEL,
  ALERT_STATUS_FIELD, ALERT_TACTIC_FIELD, ALERT_TECHNIQUE_FIELD, ALERT_TIMESTAMP_FIELD
} from '../../../../../shared/constants/alert/alert-field.constant';
import {IncidentOriginTypeEnum} from '../../../../../shared/enums/incident-response/incident-origin-type.enum';
import {AlertTags} from '../../../../../shared/types/alert/alert-tag.type';
import {AlertStatusEnum, UtmAlertType} from '../../../../../shared/types/alert/utm-alert.type';
import {ElasticFilterType} from '../../../../../shared/types/filter/elastic-filter.type';
import {IncidentCommandType} from '../../../../../shared/types/incident/incident-command.type';
import {UtmFieldType} from '../../../../../shared/types/table/utm-field.type';
import {AlertUpdateHistoryBehavior} from '../../behavior/alert-update-history.behavior';
import {AlertUpdateTagBehavior} from '../../behavior/alert-update-tag.behavior';
import {AlertHistoryActionEnum} from '../../enums/alert-history-action.enum';
import {EventDataTypeEnum} from '../../enums/event-data-type.enum';
import {AlertFieldService} from '../../services/alert-field.service';

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
  @Input() timeFilter: ElasticFilterType;
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
  ALERT_TECHNIQUE_FIELD = ALERT_TECHNIQUE_FIELD;
  IMPACT_FIELD = ALERT_IMPACT_FIELD;
  countRelatedEvents: number;
  viewHistory = false;
  viewLog = false;
  viewOnMap = true;
  alertActionEnum = AlertHistoryActionEnum;
  logs: any[];
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

  constructor(private alertUpdateHistoryBehavior: AlertUpdateHistoryBehavior,
              private alertUpdateTagBehavior: AlertUpdateTagBehavior,
              private moduleService: UtmModulesService,
              private alertFieldService: AlertFieldService) {
  }

  ngOnInit() {
    this.viewLog = this.fullScreen;
    this.logs = this.alert.events;
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

  getFieldByName(name: string): UtmFieldType {
    return this.alertFieldService.findField(ALERT_FIELDS, name);
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

  showMap(): boolean {
    return (this.alert && this.alert.target && this.alert.target.geolocation && 'latitude' in this.alert.target.geolocation && 'longitude' in this.alert.target.geolocation) &&
      (this.alert && this.alert.adversary && this.alert.adversary.geolocation && 'latitude' in this.alert.adversary.geolocation && 'longitude' in this.alert.adversary.geolocation);
  }

  areCoordinatesNotZero(coordinates: number[]): boolean {
    return coordinates[0] !== 0.0 || coordinates[1] !== 0.0;
  }


  /*searchLastLog() {
    if (!this.log) {
      this.getLastLog = true;
      const filters: ElasticFilterType[] = [
        ...(this.timeFilter ? [this.timeFilter] : []),
        {field: ALERT_ID_FIELD, operator: ElasticOperatorsEnum.IS, value: this.logs[0]}
      ];
      this.elasticDataService.search(1, 1,
        1, LOG_INDEX_PATTERN, filters).subscribe(
        (res: HttpResponse<any>) => {
          this.log = res.body[0] ? res.body[0] : {};
          this.getLastLog = false;
        },
        (res: HttpResponse<any>) => {
          this.toastService.showError('Error', 'An error occurred while trying to retrieve the latest log.');
        }
      );
    }
  }*/

  searchLastLog() {
    this.log = this.alert.events.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())[0];
  }

  viewLastLog() {
    console.log(this.alert.events);
    this.searchLastLog();
  }

  isEmptyResponse() {
    return Object.entries(this.log).length === 0;
  }

  protected readonly ALERT_ADVERSARY_FIELD = ALERT_ADVERSARY_FIELD;
}

export enum AlertDetailTabEnum {
  DETAIL = 'detail',
  SOC_AI = 'socAi',
  HISTORY = 'history',
  LAST_LOG = 'lastLog',
  MAP = 'map',
  TAGS = 'tags',
  INCIDENT = 'incident',
  ECHOES = 'echoes',
}
