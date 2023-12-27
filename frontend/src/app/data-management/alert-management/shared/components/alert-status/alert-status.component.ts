import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {TranslateService} from '@ngx-translate/core';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
import {AlertIncidentStatusChangeBehavior} from '../../../../../shared/behaviors/alert-incident-status-change.behavior';
import {
  AUTOMATIC_REVIEW,
  CLOSED,
  CLOSED_ICON,
  IGNORED,
  IGNORED_ICON,
  OPEN,
  OPEN_ICON,
  REVIEW,
  REVIEW_ICON
} from '../../../../../shared/constants/alert/alert-status.constant';
import {UtmIncidentAlertsService} from '../../../../../shared/services/incidents/utm-incident-alerts.service';
import {UtmAlertType} from '../../../../../shared/types/alert/utm-alert.type';
import {AlertIncidentStatusUpdateType} from '../../../../../shared/types/incident/alert-incident-status-update.type';
import {AlertStatusBehavior} from '../../behavior/alert-status.behavior';
import {AlertUpdateHistoryBehavior} from '../../behavior/alert-update-history.behavior';
import {EventDataTypeEnum} from '../../enums/event-data-type.enum';
import {AlertManagementService} from '../../services/alert-management.service';
import {getStatusName} from '../../util/alert-util-function';
import {AlertCompleteComponent} from '../alert-complete/alert-complete.component';

@Component({
  selector: 'app-alert-status',
  templateUrl: './alert-status.component.html',
  styleUrls: ['./alert-status.component.scss']
})
export class AlertStatusComponent implements OnInit {
  @Input() alert: UtmAlertType;
  @Input() showDrop: boolean;
  @Input() statusField: string;
  @Output() statusChange = new EventEmitter<number>();
  @Input() status: number;
  @Input() dataType: EventDataTypeEnum;
  eventDataTypeEnum = EventDataTypeEnum;
  icon: string;
  label: string;
  background: string;
  open = OPEN;
  review = REVIEW;
  ignored = IGNORED;
  closed = CLOSED;
  openIcon = OPEN_ICON;
  reviewIcon = REVIEW_ICON;
  ignoredIcon = IGNORED_ICON;
  closedIcon = CLOSED_ICON;
  pending = AUTOMATIC_REVIEW;
  changing = false;
  isIncident: boolean;
  incidentId: number;

  constructor(private alertServiceManagement: AlertManagementService,
              private modalService: NgbModal,
              private translate: TranslateService,
              private alertUpdateHistoryBehavior: AlertUpdateHistoryBehavior,
              private updateStatusServiceBehavior: AlertStatusBehavior,
              private alertIncidentStatusChangeBehavior: AlertIncidentStatusChangeBehavior,
              private utmIncidentAlertsService: UtmIncidentAlertsService,
              private utmToastService: UtmToastService) {
  }

  ngOnInit() {
    if (typeof this.status === 'string') {
      this.status = Number(this.status);
    }
    if (!this.status) {
      this.status = this.alert.status;
    }
    this.isIncident = this.alert.isIncident;
    if (this.isIncident) {
      this.incidentId = Number(this.alert.incidentDetail.incidentId);
    }
    // if (this.showDrop) {
    //   this.showDrop = !(this.status === this.pending || this.status === this.ignored || this.status === this.closed);
    // }
    this.resolveAlert();
  }

  changeStatus(status: number, alert: UtmAlertType) {
    this.changing = true;
    if (status === this.closed) {
      const modalRef = this.modalService.open(AlertCompleteComponent, {centered: true});
      modalRef.componentInstance.alertsIDs = [alert.id];
      modalRef.componentInstance.canCreateRule = true;
      modalRef.componentInstance.status = status;
      modalRef.componentInstance.alert = this.alert;
      modalRef.componentInstance.statusClose.subscribe(() => {
        this.changing = false;
      });
      modalRef.componentInstance.statusChange.subscribe((statusChange) => {
        this.changing = false;
        if (statusChange === 'success') {
          this.statusChange.emit(status);
          this.alertUpdateHistoryBehavior.$refreshHistory.next(true);
          this.statusChangedSuccess(status);
          this.syncIncidentAlertStatus([this.alert.id], status);
        } else {
          this.changing = false;
        }
      });
    } else {
      this.alertServiceManagement.updateAlertStatus([this.alert.id], status).subscribe(al => {
        this.statusChangedSuccess(status);
        this.changing = false;
        this.alertUpdateHistoryBehavior.$refreshHistory.next(true);
        this.syncIncidentAlertStatus([this.alert.id], status);
      });
    }
  }

  syncIncidentAlertStatus(alerts: string[], status: number) {
    if (this.isIncident) {
      const update: AlertIncidentStatusUpdateType = {
        status,
        incidentId: Number(this.incidentId),
        alertIds: alerts
      };
      this.utmIncidentAlertsService.updateIncidentAlertStatus(update).subscribe(() => {
        this.alertIncidentStatusChangeBehavior.$incidentAlertChange.next(this.incidentId);
      });
    }
  }

  statusChangedSuccess(status) {
    this.updateStatusServiceBehavior.$updateStatus.next(true);
    this.status = status;
    this.resolveAlert();
    this.statusChange.emit(status);
    const msg = getStatusName(status);
    this.translate.get(['toast.changeAlertStatus', msg]).subscribe(value => {
      this.utmToastService.showSuccessBottom(value['toast.changeAlertStatus'] + ' ' + value[msg].toString().toUpperCase());
    });
  }

  private resolveAlert() {
    switch (this.status) {
      case OPEN:
        this.icon = OPEN_ICON;
        this.background = 'border-success-400 text-success-400';
        this.label = 'alertStatus.open';
        break;
      case REVIEW:
        this.icon = REVIEW_ICON;
        this.background = 'border-info-400 text-info-400';
        this.label = 'alertStatus.inReview';
        break;
      case CLOSED:
        this.icon = CLOSED_ICON;
        this.background = 'border-blue-800 text-blue-800';
        this.label = 'alertStatus.closed';
        break;
      case IGNORED:
        this.icon = IGNORED_ICON;
        this.background = 'border-warning-400 text-warning-400';
        this.label = 'alertStatus.ignored';
        break;
      default:
        this.icon = 'icon-hammer';
        this.background = 'border-slate-800 text-slate-800';
        this.label = 'alertStatus.pending';
        break;
    }
  }
}
