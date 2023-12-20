import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {User} from '../../../../core/user/user.model';
import {UserService} from '../../../../core/user/user.service';
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
import {PrefixElementEnum} from '../../../../shared/enums/prefix-element.enum';
import {UtmIncidentService} from '../../../../shared/services/incidents/utm-incident.service';
import {NewIncidentAlert} from '../../../../shared/types/incident/new-incident.type';
import {UtmIncidentType} from '../../../../shared/types/incident/utm-incident.type';
import {getValueFromPropertyPath} from '../../../../shared/util/get-value-object-from-property-path.util';
import {InputClassResolve} from '../../../../shared/util/input-class-resolve';
import {createElementPrefix} from '../../../../shared/util/string-util';

@Component({
  selector: 'app-create-incident',
  templateUrl: './create-incident.component.html',
  styleUrls: ['./create-incident.component.scss']
})
export class CreateIncidentComponent implements OnInit {
  @Input() alerts: any[];
  @Output() incidentAdded = new EventEmitter<UtmIncidentType>();
  alertList: NewIncidentAlert[];
  users: User[];
  formIncident: FormGroup;
  step = 1;
  stepCompleted: number[] = [];
  totalItems: number;
  page = 1;
  itemsPerPage = 5;
  pageStart = 0;
  pageEnd = 5;
  creating = false;
  irPrefix = createElementPrefix(PrefixElementEnum.INCIDENT);

  constructor(private userService: UserService,
              private incidentService: UtmIncidentService,
              public inputClassResolve: InputClassResolve,
              private activeModal: NgbActiveModal,
              private alertManagementService: AlertManagementService,
              private utmToastService: UtmToastService,
              private alertUpdateHistoryBehavior: AlertUpdateHistoryBehavior,
              private fb: FormBuilder) {

  }


  ngOnInit() {
    this.getUsers();
    this.initForm();
    if (this.alerts.length > 0) {
      this.totalItems = this.alerts.length;
      this.convertAlertsToIncidentAlerts(this.alerts).then((incidentAlerts) => {
        this.alertList = incidentAlerts;
        this.formIncident.controls.alertList.setValue(this.alertList);
      });
    }
  }

  initForm() {
    this.formIncident = this.fb.group({
      incidentName: ['', [Validators.required]],
      incidentDescription: ['', [Validators.required]],
      incidentAssignedTo: [],
      alertList: [null, [Validators.required]]
    });
  }

  getUsers() {
    this.userService.query({size: 1000}).subscribe((res) => {
      this.users = res.body;
    });
  }

  nextStep() {
    this.stepCompleted.push(this.step);
    this.step += 1;
  }

  isCompleted(step: number) {
    return this.stepCompleted.findIndex(value => value === step) !== -1;
  }

  backStep() {
    this.stepCompleted.pop();
    this.step -= 1;
  }

  loadPage(page: number) {
    this.pageEnd = page * this.itemsPerPage;
    this.pageStart = this.pageEnd - this.itemsPerPage;
  }

  createIncident() {
    this.creating = true;
    const value = this.formIncident.get('incidentName').value;
    this.formIncident.get('incidentName').setValue(this.irPrefix + value);
    this.incidentService.create(this.formIncident.value).subscribe((res) => {
      this.addAsIncident(res.body);
      this.utmToastService.showSuccessBottom('Incident created successfully');
    }, () => {
      this.creating = false;
      const ruleName: string = this.formIncident.get('incidentName').value;
      this.formIncident.get('incidentName').setValue(ruleName.replace(this.irPrefix, ''));
      this.utmToastService.showError('Error creating incident',
        'Han error occurred while creating incident, please contact your administrator');
    });

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

  onUserChange($event: any) {
    this.formIncident.controls.incidentAssignedTo.setValue($event);
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

  setUsersAssignments($event: User[]) {
    const ids = $event.map(user => {
      return {id: user.id, login: user.login};
    });
    this.formIncident.controls.incidentAssignedTo.setValue(JSON.stringify(ids));
  }
}
