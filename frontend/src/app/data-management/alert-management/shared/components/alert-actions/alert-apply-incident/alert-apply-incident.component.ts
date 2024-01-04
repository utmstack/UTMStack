import {Component, EventEmitter, Input, OnInit, Output, ViewChild} from '@angular/core';
import {NgbModal, NgbPopover} from '@ng-bootstrap/ng-bootstrap';
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
  @ViewChild('incidentPopoverSpan') incidentPopoverSpan: NgbPopover;
  @ViewChild('incidentPopoverButton') incidentPopoverButton: NgbPopover;
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

  checkAlerts() {
    if (this.multiple && this.alerts && this.alerts.length > 0) {
      const nonIncidentAlerts = this.alerts.filter(({ isIncident }) => !isIncident);
      if (nonIncidentAlerts.length > 0) {
        return nonIncidentAlerts;
      } else {
        this.utmToastService.showError(this.alerts && this.alerts.length > 1 ? 'Alerts associated with incident'
            : 'Alert associated with incident',
          this.alerts && this.alerts.length > 1 ? 'The selected alerts are associated with an incident.'
            : 'The selected alert is associated with an incident.');
        return [];
      }
    } else {
      return [this.alert];
    }
  }

  addToIncident() {
    const alerts = this.checkAlerts();
    if (alerts.length > 0) {
      const modal = this.modalService.open(AddAlertToIncidentComponent, {size: 'lg', backdrop: 'static', centered: true});
      modal.componentInstance.alerts = alerts;

      modal.componentInstance.incidentAdded.subscribe((incident) => {
        this.markAsIncident.emit(incident.id.toString());
      });
    }
  }

  createIncident() {
    const alerts = this.checkAlerts();
    if (alerts.length > 0) {
      const modal = this.modalService.open(CreateIncidentComponent, {size: 'lg', backdrop: 'static', centered: true});
      modal.componentInstance.alerts = alerts;

      modal.componentInstance.incidentAdded.subscribe((incident) => {
        this.markAsIncident.emit(incident.id.toString());
      });
    }
  }

  closePopover() {
    if (this.incidentPopoverSpan) {
      this.incidentPopoverSpan.close();
    }

    if (this.incidentPopoverButton) {
      this.incidentPopoverButton.close();
    }
  }

}
