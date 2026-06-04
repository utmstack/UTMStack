import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NetScanType} from '../../../../assets-discover/shared/types/net-scan.type';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {IncidentOriginTypeEnum} from '../../../../shared/enums/incident-response/incident-origin-type.enum';
import {IncidentResponseActionsEnum} from '../../../../shared/enums/incident-response/incident-response-actions.enum';
import {IncidentResponseStatusEnum} from '../../../../shared/enums/incident-response/incident-response-status.enum';
import {IncidentResponseAssetService} from '../../services/incident-response-asset.service';
import {IncidentResponseJobService} from '../../services/incident-response-job.service';
import {IncidentActionType} from '../../type/incident-action.type';
import {IncidentJobType} from '../../type/incident-job.type';

@Component({
  selector: 'app-ir-execute-command',
  templateUrl: './ir-execute-command.component.html',
  styleUrls: ['./ir-execute-command.component.scss']
})
export class IrExecuteCommandComponent implements OnInit {
  @Input() originType: IncidentOriginTypeEnum = IncidentOriginTypeEnum.USER_EXECUTION;
  @Input() originId = 0;
  utmAssets: NetScanType[];
  asset: any;
  creating = false;
  action: IncidentActionType;
  irActionsEnum = IncidentResponseActionsEnum;
  param: string;
  ip: string;
  @Output() jobCreated = new EventEmitter<string>();


  constructor(private incidentResponseAssetService: IncidentResponseAssetService,
              public utmToastService: UtmToastService,
              private incidentResponseJobService: IncidentResponseJobService) {
  }

  ngOnInit() {
    this.getAssetList();
  }

  getAssetList() {
    this.incidentResponseAssetService.query({size: 1000, isAlive: true}).subscribe(response => {
      this.utmAssets = response.body;
    }, error => {
      this.utmToastService.showError('Error', 'Error loading assets list');
    });
  }

  onSelectChange($event: any) {
  }

  addJob() {
    const job: IncidentJobType = {
      actionId: this.action.id,
      params: this.param,
      agent: this.asset,
      status: IncidentResponseStatusEnum.PENDING,
      originType: this.originType,
      originId: this.originId,
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
    if (!this.action || !this.asset) {
      return false;
    } else {
      return !((this.action.actionType !== this.irActionsEnum.ISOLATE_HOST
        && this.action.actionType !== this.irActionsEnum.SHUTDOWN_SERVER
        && this.action.actionType !== this.irActionsEnum.RESTART_SERVER) && !this.param);
    }
  }

  onActionSelect($event: { action: IncidentActionType; param: string }) {
    this.action = $event.action;
    this.param = $event.param;
  }
}
