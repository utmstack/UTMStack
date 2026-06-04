import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {IncidentResponseActionsEnum} from '../../../../shared/enums/incident-response/incident-response-actions.enum';
import {IncidentResponseActionService} from '../../services/incident-response-action.service';
import {IncidentActionType} from '../../type/incident-action.type';

@Component({
  selector: 'app-ir-command-select',
  templateUrl: './ir-command-select.component.html',
  styleUrls: ['./ir-command-select.component.scss']
})
export class IrCommandSelectComponent implements OnInit {
  actions: IncidentActionType[] = [];
  actionApply: IncidentActionType;
  loading = true;
  irActionsEnum = IncidentResponseActionsEnum;
  requestParams: any;
  @Input() action: IncidentActionType;
  @Output() actionSelect = new EventEmitter<{ action: IncidentActionType, param: string }>();
  param: any;

  constructor(private incidentResponseActionService: IncidentResponseActionService) {
  }

  ngOnInit() {
    if (this.action) {
      this.actionApply = this.action;
    }
    this.requestParams = {
      page: 0,
      size: 50,
      sort: 'actionType,asc',
      'actionCommand.contains': ''
    };
    this.getActions();
  }

  getActions() {
    this.incidentResponseActionService.query(this.requestParams).subscribe(response => {
      this.actions = response.body;
      this.loading = false;
    });
  }

  getLabelParam(): string {
    let label = '';
    switch (this.actionApply.actionType) {
      case IncidentResponseActionsEnum.DISABLE_USER:
        label = 'User to disable';
        break;
      case IncidentResponseActionsEnum.BLOCK_IP:
        label = 'IP to block';
        break;
      case IncidentResponseActionsEnum.KILL_PROCESS:
        label = 'Process name to kill';
        break;
      case IncidentResponseActionsEnum.UNINSTALL_PROGRAM:
        label = 'Program to uninstall';
        break;
    }
    return label;
  }

  select(command: IncidentActionType) {
    this.actionApply = command;
    this.param = undefined;
    this.actionSelect.emit({action: command, param: this.param});
  }
}
