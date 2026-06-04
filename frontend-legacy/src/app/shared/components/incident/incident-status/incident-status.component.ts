import {Component, EventEmitter, Input, OnChanges, OnInit, Output, SimpleChanges} from '@angular/core';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../alert/utm-toast.service';
import {AlertIncidentStatusChangeBehavior} from '../../../behaviors/alert-incident-status-change.behavior';
import {IncidentStatusEnum} from '../../../enums/incident/incident-status.enum';
import {UtmIncidentService} from '../../../services/incidents/utm-incident.service';
import {UtmIncidentType} from '../../../types/incident/utm-incident.type';
import {
  INCIDENT_COMPLETE_ICON,
  INCIDENT_COMPLETE_LABEL,
  INCIDENT_OPEN_ICON,
  INCIDENT_OPEN_LABEL,
  INCIDENT_REVIEW_ICON,
  INCIDENT_REVIEW_LABEL
} from '../../../../incident/incident-shared/const/incident-status.constant';

@Component({
  selector: 'app-modal-complete-incident',
  template: `
    <app-utm-modal-header [name]="'Complete incident'" class="w-100"
                          (closeModal)="activeModal.close()"></app-utm-modal-header>
    <div class="container-fluid p-3">
      <div
        class="alert bg-warning-400 text-white alert-styled-right">
        <span class="font-weight-semibold">Warning! </span>
        <span>Enter incident solution or steps you do to complete this incident</span>
      </div>
      <div class="observations-container d-flex flex-column">
        <label class="" for="obs">Solution ({{1000 - (incident.incidentSolution ? incident.incidentSolution.length : 0)}})</label>
        <textarea [(ngModel)]=" incident.incidentSolution" class="border-1 border-grey-600 form-control" id="obs" name=""
                  rows="5"></textarea>

        <div class="button-container d-flex justify-content-end mt-3">
          <button (click)="activeModal.close();"
                  class="btn utm-button utm-button-grey mr-3">
            <i class="icon-close2"></i>&nbsp;Cancel
          </button>
          <button (click)="changeStatus()" [disabled]="incident.incidentSolution==='' || !incident.incidentSolution"
                  class="btn utm-button utm-button-primary">
            <i [ngClass]="'icon-checkmark-circle'"></i>&nbsp;
            Apply status
          </button>
        </div>
      </div>
    </div>
  `,
})
export class CompleteIncidentModalComponent {
  @Input() incident: UtmIncidentType;
  @Output() incidentStatusChanged = new EventEmitter<UtmIncidentType>();

  constructor(public activeModal: NgbActiveModal) {
  }

  changeStatus() {
    this.incident.incidentStatus = IncidentStatusEnum.COMPLETED;
    this.activeModal.close();
    this.incidentStatusChanged.emit(this.incident);
  }
}

@Component({
  selector: 'app-incident-status',
  templateUrl: './incident-status.component.html',
  styleUrls: ['./incident-status.component.css']
})
export class IncidentStatusComponent implements OnInit, OnChanges {
  @Input() incident: UtmIncidentType;
  @Output() statusChange = new EventEmitter<IncidentStatusEnum>();
  background: string;
  icon: string;
  label: string;
  changing = false;
  incidentStatusEnum = IncidentStatusEnum;
  open = INCIDENT_OPEN_LABEL;
  review = INCIDENT_REVIEW_LABEL;
  complete = INCIDENT_COMPLETE_LABEL;
  openIcon = INCIDENT_OPEN_ICON;
  reviewIcon = INCIDENT_REVIEW_ICON;
  completeIcon = INCIDENT_COMPLETE_ICON;

  constructor(private utmIncidentService: UtmIncidentService,
              private modalService: NgbModal,
              private alertIncidentStatusChangeBehavior: AlertIncidentStatusChangeBehavior,
              private toastService: UtmToastService) {
  }

  ngOnInit() {
    this.resolveIncidentStatus();
  }

  ngOnChanges(changes: SimpleChanges) {
  }

  private resolveIncidentStatus() {
    switch (this.incident.incidentStatus) {
      case IncidentStatusEnum.OPEN:
        this.icon = INCIDENT_OPEN_ICON;
        this.background = 'border-success-400 text-success-400';
        this.label = INCIDENT_OPEN_LABEL;
        break;
      case IncidentStatusEnum.IN_REVIEW:
        this.icon = INCIDENT_REVIEW_ICON;
        this.background = 'border-info-400 text-info-400';
        this.label = INCIDENT_REVIEW_LABEL;
        break;
      case IncidentStatusEnum.COMPLETED:
        this.icon = INCIDENT_COMPLETE_ICON;
        this.background = 'border-blue-800 text-blue-800';
        this.label = INCIDENT_COMPLETE_LABEL;
        break;
      default:
        this.icon = 'icon-hammer';
        this.background = 'border-slate-800 text-slate-800';
        this.label = 'NOT FOUND';
        break;
    }
  }

  changeStatus(status: IncidentStatusEnum) {
    this.changing = true;
    if (status !== IncidentStatusEnum.COMPLETED) {
      this.incident.incidentStatus = status;
      this.utmIncidentService.changeStatus(this.incident).subscribe(() => {
        this.statusModified(status);
      });
    } else {
      const modal = this.modalService.open(CompleteIncidentModalComponent, {centered: true});
      modal.componentInstance.incident = this.incident;
      modal.componentInstance.incidentStatusChanged.subscribe((incident: UtmIncidentType) => {
        this.utmIncidentService.changeStatus(incident).subscribe(() => {
          this.incident.incidentStatus = status;
          this.statusModified(status);
        });
      });
    }
  }

  statusModified(status: IncidentStatusEnum) {
    this.toastService.showSuccessBottom('Incident and related alert status change to ' + status);
    this.changing = false;
    this.statusChange.emit(status);
    this.alertIncidentStatusChangeBehavior.$incidentAlertChange.next(this.incident.id);
    this.resolveIncidentStatus();
  }
}
