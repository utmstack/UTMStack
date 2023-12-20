import {Component, Input, OnInit} from '@angular/core';
import {IncidentResponseStatusEnum} from '../../../../shared/enums/incident-response/incident-response-status.enum';

@Component({
  selector: 'app-incident-response-status',
  templateUrl: './incident-response-status.component.html',
  styleUrls: ['./incident-response-status.component.scss']
})
export class IncidentResponseStatusComponent implements OnInit {
  @Input() status: IncidentResponseStatusEnum;
  irStatusEnum = IncidentResponseStatusEnum;
  background: string;
  label: string;

  constructor() {
  }

  ngOnInit() {
    this.setBackgroundColor();
  }

  private setBackgroundColor() {
    switch (this.status) {
      case IncidentResponseStatusEnum.ERROR:
        this.background = 'border-danger-800 text-danger-800';
        this.label = 'Error';
        break;
      case IncidentResponseStatusEnum.EXECUTED:
        this.background = 'border-success-800 text-success-800';
        this.label = 'Executed';
        break;
      case IncidentResponseStatusEnum.PENDING:
        this.background = 'border-slate-800 text-slate-800';
        this.label = 'Pending';
        break;
      case IncidentResponseStatusEnum.RUNNING:
        this.label = 'Pending';
        this.background = 'border-info-800 text-info-800';
        break;
    }
  }

}
