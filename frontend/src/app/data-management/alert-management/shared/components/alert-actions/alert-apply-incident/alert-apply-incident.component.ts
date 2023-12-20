import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {
  AddAlertToIncidentComponent
} from '../../../../../../incident/incident-shared/component/add-alert-to-incident/add-alert-to-incident.component';
import {CreateIncidentComponent} from '../../../../../../incident/incident-shared/component/create-incident/create-incident.component';
import {UtmToastService} from '../../../../../../shared/alert/utm-toast.service';
import {UtmAlertType} from '../../../../../../shared/types/alert/utm-alert.type';
import {AlertUpdateHistoryBehavior} from '../../../behavior/alert-update-history.behavior';
import {EventDataTypeEnum} from '../../../enums/event-data-type.enum';
import {AlertManagementService} from '../../../services/alert-management.service';

@Component({
  selector: 'app-alert-apply-incident',
  templateUrl: './alert-apply-incident.component.html',
  styleUrls: ['./alert-apply-incident.component.scss']
})
export class AlertApplyIncidentComponent implements OnInit {
  @Input() alert: UtmAlertType;
  /**
   * Selected id in tables
   */
  @Input() alerts: any[];
  @Input() multiple = false;
  @Input() eventType: EventDataTypeEnum;
  @Output() markAsIncident = new EventEmitter<string>();
  creating: any;
  isIncident: boolean;
  incidentName: string;

  constructor(private alertManagementService: AlertManagementService,
              private utmToastService: UtmToastService,
              private modalService: NgbModal,
              private alertUpdateHistoryBehavior: AlertUpdateHistoryBehavior) {
  }

  ngOnInit() {
    if (!this.multiple && this.alert) {
      this.isIncident = this.alert.isIncident;
      this.incidentName = this.alert.incidentDetail ? this.alert.incidentDetail.incidentName : '';
    }
  }


  createIncident() {
    const modal = this.modalService.open(CreateIncidentComponent, {size: 'lg', backdrop: 'static', centered: true});
    if (this.multiple && this.alerts && this.alerts.length > 0) {
      modal.componentInstance.alerts = this.alerts;
    } else {
      modal.componentInstance.alerts = [this.alert];
    }
    modal.componentInstance.incidentAdded.subscribe((incident) => {
      this.markAsIncident.emit(incident.id.toString());
    });
  }

  addToIncident() {
    const modal = this.modalService.open(AddAlertToIncidentComponent, {size: 'lg', backdrop: 'static', centered: true});
    if (this.multiple && this.alerts && this.alerts.length > 0) {
      modal.componentInstance.alerts = this.alerts;
    } else {
      modal.componentInstance.alerts = [this.alert];
    }
    modal.componentInstance.incidentAdded.subscribe((incident) => {
      this.markAsIncident.emit(incident.id.toString());
    });
  }
}
