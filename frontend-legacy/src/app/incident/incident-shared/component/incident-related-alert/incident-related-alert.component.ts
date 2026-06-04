import {Component, Input, OnInit} from '@angular/core';
import {EventDataTypeEnum} from '../../../../data-management/alert-management/shared/enums/event-data-type.enum';
import {AlertTagService} from '../../../../data-management/alert-management/shared/services/alert-tag.service';
import {AlertIncidentStatusChangeBehavior} from '../../../../shared/behaviors/alert-incident-status-change.behavior';
import {ALERT_CASE_ID_FIELD} from '../../../../shared/constants/alert/alert-field.constant';
import {SortEvent} from '../../../../shared/directives/sortable/type/sort-event';
import {ElasticOperatorsEnum} from '../../../../shared/enums/elastic-operators.enum';
import {DataNatureTypeEnum} from '../../../../shared/enums/nature-data.enum';
import {ElasticDataService} from '../../../../shared/services/elasticsearch/elastic-data.service';
import {UtmIncidentAlertsService} from '../../../../shared/services/incidents/utm-incident-alerts.service';
import {AlertTags} from '../../../../shared/types/alert/alert-tag.type';
import {IncidentAlertType} from '../../../../shared/types/incident/incident-alert.type';

@Component({
  selector: 'app-incident-related-alert',
  templateUrl: './incident-related-alert.component.html',
  styleUrls: ['./incident-related-alert.component.css']
})
export class IncidentRelatedAlertComponent implements OnInit {
  @Input() incidentId: number;
  request = {
    page: 0,
    size: 10,
    sort: 'id,desc',
    'incidentId.equals': this.incidentId
  };
  alerts: IncidentAlertType[];
  totalItems: number;
  loading = true;
  tags: AlertTags[];
  alert: any;
  loadingAlert = false;
  viewAlertDetail = false;
  alertSelected: IncidentAlertType;
  eventDataTypeEnum = EventDataTypeEnum;

  constructor(private incidentAlertService: UtmIncidentAlertsService,
              private elasticDataService: ElasticDataService,
              private alertIncidentStatusChangeBehavior: AlertIncidentStatusChangeBehavior,
              private alertTagService: AlertTagService) {
  }

  ngOnInit() {
    this.request['incidentId.equals'] = this.incidentId;
    this.getIncidentAlerts();
    this.getTags();
    this.alertIncidentStatusChangeBehavior.$incidentAlertChange.subscribe(incidentId => {
      this.getIncidentAlerts();
    });
  }

  getIncidentAlerts() {
    this.incidentAlertService.query(this.request).subscribe((res) => {
      this.alerts = res.body;
      this.totalItems = Number(res.headers.get('X-Total-Count'));
      this.loading = false;
    });
  }

  loadPage(page: number) {
    this.request.page = page - 1;
    this.getIncidentAlerts();
  }

  onSortBy($event: SortEvent) {
    this.request.sort = $event.column + ',' + $event.direction;
    this.getIncidentAlerts();
  }

  getTags() {
    this.alertTagService.query({page: 0, size: 1000}).subscribe(reponse => {
      this.tags = reponse.body;
    });
  }

  searchAlert(alertId: string) {
    const filterAlert = [{
      field: ALERT_CASE_ID_FIELD,
      operator: ElasticOperatorsEnum.IS,
      value: alertId
    }];
    this.elasticDataService.search(1, 1, 1, DataNatureTypeEnum.ALERT, filterAlert)
      .subscribe(reponse => {
        this.alert = reponse.body[0];
        this.loadingAlert = false;
      });
  }

  showAlertDetail(alert: IncidentAlertType) {
    this.viewAlertDetail = true;
    this.loadingAlert = true;
    this.alertSelected = alert;
    this.searchAlert(alert.alertId);
  }

  onRefreshData($event: boolean) {

  }
}
