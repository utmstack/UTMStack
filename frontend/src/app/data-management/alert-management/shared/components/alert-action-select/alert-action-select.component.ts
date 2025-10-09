import {Component, EventEmitter, Input, OnDestroy, OnInit, Output} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {TranslateService} from '@ngx-translate/core';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
import {AlertIncidentStatusChangeBehavior} from '../../../../../shared/behaviors/alert-incident-status-change.behavior';
import {
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
import {AlertStatusEnum, UtmAlertType} from '../../../../../shared/types/alert/utm-alert.type';
import {AlertIncidentStatusUpdateType} from '../../../../../shared/types/incident/alert-incident-status-update.type';
import {AlertStatusBehavior} from '../../behavior/alert-status.behavior';
import {AlertUpdateHistoryBehavior} from '../../behavior/alert-update-history.behavior';
import {EventDataTypeEnum} from '../../enums/event-data-type.enum';
import {AlertManagementService} from '../../services/alert-management.service';
import {getStatusName} from '../../util/alert-util-function';
import {AlertCompleteComponent} from '../alert-complete/alert-complete.component';

export enum AlertActionType {
  ADD_OR_CREATE_INCIDENT,
  ADD_FALSE_POSITIVE_TAG_RULE,
  ADD_INCIDENT,
  CREATE_INCIDENT
}

export interface AlertAction {
  label: string;
  value?: AlertStatusEnum;
  group: string;
  subActions?: AlertAction[];
  icon?: string;
  background?: string;
  action?: AlertActionType;
}

@Component({
  selector: 'app-alert-action-select',
  templateUrl: './alert-action-select.component.html',
  styleUrls: ['./alert-action-select.component.css']
})
export class AlertActionSelectComponent implements OnInit, OnDestroy {

  @Input() alert: UtmAlertType;
  @Input() showDrop: boolean;
  @Input() statusField: string;
  @Output() statusChange = new EventEmitter<number>();
  @Input() status: number;
  @Input() tags: any[];
  @Input() dataType: EventDataTypeEnum;


  rawActions: AlertAction[] = [
    { label: 'Open', value: AlertStatusEnum.OPEN , group: 'Estado', icon: OPEN_ICON, background: 'border-success-400 text-success-400' },
    { label: 'In Review', value: AlertStatusEnum.IN_REVIEW, group: 'Estado',
      icon: REVIEW_ICON, background: 'border-info-400 text-info-400' },
    {label: 'Completed', value: AlertStatusEnum.COMPLETED, group: 'Estado', icon: CLOSED_ICON,
      background: 'border-blue-800 text-blue-800'},
    {label: 'Completed as False Positive', value: AlertStatusEnum.COMPLETED_AS_FALSE_POSITIVE, group: 'Estado', icon: 'icon-checkmark',
      background: 'border-warning-400 text-warning-800'},
      /*subActions: [
        { label: 'Completed', value: 'completed_plain', group: 'Estado', icon: CLOSED_ICON,
          background: 'border-blue-800 text-blue-800', },
        { label: 'Completed – Has False Positive', value: 'completed_fp', group: 'Estado', icon: CLOSED_ICON,
          background: 'border-blue-800 text-blue-800', }
      ]*/
    { label: 'Manage Incident',
      value: AlertStatusEnum.AUTOMATIC_REVIEW,
      group: 'Acciones',
      icon: 'icon-target',
      background: 'border-blue-800 text-blue-800',
      action: AlertActionType.ADD_OR_CREATE_INCIDENT,
      subActions: [
        { label: 'Add to Incident', group: 'Estado', icon: 'icon-make-group',
          background: 'border-blue-800 text-blue-800', action: AlertActionType.ADD_INCIDENT },
        { label: 'Create an Incident', group: 'Estado', icon: 'icon-plus2',
          background: 'border-blue-800 text-blue-800', action: AlertActionType.CREATE_INCIDENT },
      ]},
    { label: 'Create False Positive Rule', value: AlertStatusEnum.AUTOMATIC_REVIEW, group: 'Acciones',
      icon: 'icon-price-tag3', background: 'border-blue-800 text-blue-800', action: AlertActionType.ADD_FALSE_POSITIVE_TAG_RULE },
  ];

  actionGroups: { [group: string]: AlertAction[] } = {};
  icon: string;
  background: string;
  label: string;
  changing = false;
  isIncident: boolean;
  incidentId: number;
  AlertActionType = AlertActionType;

  constructor(private alertServiceManagement: AlertManagementService,
              private modalService: NgbModal,
              private translate: TranslateService,
              private alertUpdateHistoryBehavior: AlertUpdateHistoryBehavior,
              private updateStatusServiceBehavior: AlertStatusBehavior,
              private alertIncidentStatusChangeBehavior: AlertIncidentStatusChangeBehavior,
              private utmIncidentAlertsService: UtmIncidentAlertsService,
              private utmToastService: UtmToastService) { }

  ngOnInit() {
    this.actionGroups = this.rawActions.reduce((acc, action) => {
      if (!acc[action.group]) { acc[action.group] = []; }
      acc[action.group].push(action);
      return acc;
    }, {} as { [group: string]: AlertAction[] });

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

    this.resolveAlert();
  }


  handleAction(action: any) {

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

  changeStatus(status: number) {
    const alert = this.alert;
    this.changing = true;
    if (status === AlertStatusEnum.COMPLETED || status === AlertStatusEnum.COMPLETED_AS_FALSE_POSITIVE) {
        console.log('status', status);
        const modalRef = this.modalService.open(AlertCompleteComponent, {centered: true});
        modalRef.componentInstance.alertsIDs = [alert.id];
        modalRef.componentInstance.canCreateRule = true;
        modalRef.componentInstance.status = AlertStatusEnum.COMPLETED;
        modalRef.componentInstance.asFalsePositive = status === AlertStatusEnum.COMPLETED_AS_FALSE_POSITIVE;
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

  ngOnDestroy(): void {
  }
}
