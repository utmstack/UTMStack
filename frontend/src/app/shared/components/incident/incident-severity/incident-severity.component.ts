import {ChangeDetectionStrategy, Component, Input} from '@angular/core';
import {IncidentSeverityEnum} from '../../../enums/incident/incident-severity.enum';

@Component({
  selector: 'app-incident-severity',
  templateUrl: './incident-severity.component.html',
  styleUrls: ['./incident-severity.component.css'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class IncidentSeverityComponent {
  label: string;
  background: string;

  constructor() {
  }

  @Input()
  set severity(severity: IncidentSeverityEnum) {
    switch (severity) {
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
      case IncidentSeverityEnum.UNIMPORTANT:
        this.background = 'border-green-800 text-green-800';
        this.label = 'UNIMPORTANT';
        break;
      default:
        this.background = 'border-slate-800 text-slate-800';
        this.label = 'UNKNOWN';
        break;
    }

  }

}
