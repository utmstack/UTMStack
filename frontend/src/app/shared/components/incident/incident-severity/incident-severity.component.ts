import {Component, Input, OnInit} from '@angular/core';
import {IncidentSeverityEnum} from '../../../enums/incident/incident-severity.enum';

@Component({
  selector: 'app-incident-severity',
  templateUrl: './incident-severity.component.html',
  styleUrls: ['./incident-severity.component.css']
})
export class IncidentSeverityComponent implements OnInit {
  @Input() severity: IncidentSeverityEnum;
  label: string;
  background: string;

  constructor() {
  }

  ngOnInit() {
    this.setBackgroundColor();
  }

  private setBackgroundColor() {
    switch (this.severity) {
      case IncidentSeverityEnum.HIGH:
        this.background = 'border-danger-800 text-danger-800';
        this.label = 'HIGH';
        break;
      case IncidentSeverityEnum.MEDIUM:
        this.background = 'border-warning-800 text-warning-800';
        this.label = 'MEDIUM';
        break;
      case IncidentSeverityEnum.LOW:
        this.background = 'border-info-800 text-info-800';
        this.label = 'LOW';
        break;
      default:
        this.background = 'border-slate-800 text-slate-800';
        this.label = 'UNKNOWN';
        break;
    }

  }

}
