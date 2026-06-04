import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {IncidentOriginTypeEnum} from '../../../../shared/enums/incident-response/incident-origin-type.enum';
import {IncidentResponseActionsEnum} from '../../../../shared/enums/incident-response/incident-response-actions.enum';
import {IncidentResponseStatusEnum} from '../../../../shared/enums/incident-response/incident-response-status.enum';
import {IncidentResponseActionService} from '../../services/incident-response-action.service';
import {IncidentResponseAssetService} from '../../services/incident-response-asset.service';
import {IncidentResponseJobService} from '../../services/incident-response-job.service';
import {IncidentActionType} from '../../type/incident-action.type';
import {IncidentJobType} from '../../type/incident-job.type';

@Component({
  selector: 'app-ir-job-create',
  templateUrl: './ir-job-create.component.html',
  styleUrls: ['./ir-job-create.component.scss']
})
export class IrJobCreateComponent implements OnInit {
  @Input() host: string;
  @Input() hostname: string;
  @Input() module: IncidentOriginTypeEnum;
  @Input() relatedId: string;
  @Input() action: IncidentActionType;
  @Input() param: string;
  @Input() isIcon = false;
  @Output() jobCreated = new EventEmitter<string>();

  creating = false;
  actionApply: IncidentActionType;
  cantExecute: boolean;
  irActionsEnum = IncidentResponseActionsEnum;


  constructor(private incidentResponseActionService: IncidentResponseActionService,
              private incidentResponseJobService: IncidentResponseJobService,
              public utmToastService: UtmToastService,
              private incidentResponseAssetService: IncidentResponseAssetService,
              private modalService: NgbModal) {
  }

  ngOnInit() {
    if (this.action) {
      this.actionApply = this.action;
    }
    this.incidentResponseAssetService.cantExecuteAction(this.hostname).subscribe(response => {
      this.cantExecute = response.body;
    });
  }

  addJob() {
    const job: IncidentJobType = {
      actionId: this.actionApply.id,
      params: this.param,
      agent: this.hostname,
      status: IncidentResponseStatusEnum.PENDING,
      originType: this.module,
      originId: this.relatedId,
    };
    this.creating = true;
    this.incidentResponseJobService.create(job).subscribe(response => {
      this.creating = false;
      this.utmToastService.showSuccessBottom('Job saved successfully');
      this.jobCreated.emit('success');
    }, error => {
      this.creating = false;
      this.utmToastService.showError('Error', 'Error saving job');
    });
  }


  isValid() {
    if (!this.actionApply) {
      return false;
    } else {
      return !((this.actionApply.actionType !== this.irActionsEnum.ISOLATE_HOST
        && this.actionApply.actionType !== this.irActionsEnum.SHUTDOWN_SERVER
        && this.actionApply.actionType !== this.irActionsEnum.RESTART_SERVER) && !this.param);
    }
  }

  onActionSelect($event: { action: IncidentActionType; param: string }) {
    this.actionApply = $event.action;
    this.param = $event.param;
  }
}
