import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {AlertUpdateHistoryBehavior} from '../../../../data-management/alert-management/shared/behavior/alert-update-history.behavior';
import {AlertManagementService} from '../../../../data-management/alert-management/shared/services/alert-management.service';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {
  ALERT_ID_FIELD,
  ALERT_INCIDENT_FLAG_FIELD,
  ALERT_NAME_FIELD,
  ALERT_SEVERITY_FIELD,
  ALERT_STATUS_FIELD
} from '../../../../shared/constants/alert/alert-field.constant';
import {UtmIncidentService} from '../../../../shared/services/incidents/utm-incident.service';
import {NewIncidentAlert} from '../../../../shared/types/incident/new-incident.type';
import {UtmIncidentType} from '../../../../shared/types/incident/utm-incident.type';
import {getValueFromPropertyPath} from '../../../../shared/util/get-value-object-from-property-path.util';
import {InputClassResolve} from '../../../../shared/util/input-class-resolve';

@Component({
  selector: 'app-add-alert-to-incident',
  templateUrl: './add-alert-to-incident.component.html',
  styleUrls: ['./add-alert-to-incident.component.css']
})
export class AddAlertToIncidentComponent implements OnInit {
  @Input() alerts: any[];
  @Output() incidentAdded = new EventEmitter<UtmIncidentType>();
  alertList: NewIncidentAlert[];
  formIncident: FormGroup;
  totalItems: number;
  page = 1;
  itemsPerPage = 5;
  pageStart = 0;
  pageEnd = 5;
  creating = false;
  incidents: UtmIncidentType[];

  constructor(private incidentService: UtmIncidentService,
              public inputClassResolve: InputClassResolve,
              public activeModal: NgbActiveModal,
              private alertManagementService: AlertManagementService,
              private utmToastService: UtmToastService,
              private alertUpdateHistoryBehavior: AlertUpdateHistoryBehavior,
              private fb: FormBuilder) {

  }


  ngOnInit() {
    this.initForm();
    this.getIncidents();
    if (this.alerts.length > 0) {
      this.totalItems = this.alerts.length;
      this.convertAlertsToIncidentAlerts(this.alerts).then((incidentAlerts) => {
        this.alertList = incidentAlerts;
        this.formIncident.controls.alertList.setValue(this.alertList);
      });
    }
  }

  getIncidents() {
    this.incidentService.query({size: 10000}).subscribe(response => {
      this.incidents = response.body;
    });
  }

  initForm() {
    this.formIncident = this.fb.group({
      incidentId: ['', [Validators.required]],
      alertList: [null, [Validators.required]]
    });
  }

  loadPage(page: number) {
    this.pageEnd = page * this.itemsPerPage;
    this.pageStart = this.pageEnd - this.itemsPerPage;
  }


  deleteAlert(alert: any) {
    const index = this.alertList.findIndex(value => value.alertId === alert.id);
    this.alertList.splice(index, 1);
    this.totalItems = this.alertList.length;
    this.formIncident.controls.alertList.setValue(this.alertList);
  }

  convertAlertsToIncidentAlerts(alerts: any[]): Promise<NewIncidentAlert[]> {
    return new Promise((resolve, reject) => {
      const incidentAlerts: NewIncidentAlert[] = [];
      alerts.forEach(alert => {
        const isIncident = getValueFromPropertyPath(alert, ALERT_INCIDENT_FLAG_FIELD, null);
        if (!isIncident) {
          incidentAlerts.push({
            alertId: getValueFromPropertyPath(alert, ALERT_ID_FIELD, null),
            alertName: getValueFromPropertyPath(alert, ALERT_NAME_FIELD, null),
            alertStatus: Number(getValueFromPropertyPath(alert, ALERT_STATUS_FIELD, null)),
            alertSeverity: Number(getValueFromPropertyPath(alert, ALERT_SEVERITY_FIELD, null))
          });
        }
      });
      incidentAlerts.sort((a, b) => a.alertSeverity < b.alertSeverity ? 1 : -1);
      resolve(incidentAlerts);
    });
  }

  addAsIncident(incident: UtmIncidentType) {
    this.creating = true;
    const alertIds = this.alertList.map(alert => alert.alertId);
    this.alertManagementService.markAsIncident(alertIds, incident.incidentName, incident.id, 'INCIDENT').subscribe(response => {
      this.alertUpdateHistoryBehavior.$refreshHistory.next(true);
      this.incidentAdded.emit(incident);
      this.activeModal.close();
    }, error => {
      this.utmToastService.showError('Error adding incident',
        'Error adding incident, please try again');
      this.creating = false;
    });
  }

  addAlertsToIncident() {
    this.creating = true;
    this.incidentService.addAlerts(this.formIncident.value).subscribe((res) => {
      this.addAsIncident(res.body);
      this.utmToastService.showSuccessBottom('Alerts added to Incident successfully');
    }, () => {
      this.creating = false;
      this.utmToastService.showError('Error creating incident',
        'Han error occurred while creating incident, please contact your administrator');
    });
  }
}
