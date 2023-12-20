import {Component, Input, OnInit} from '@angular/core';
import {HIGH, HIGH_TEXT, LOW, LOW_TEXT, MEDIUM, MEDIUM_TEXT} from '../../../../../shared/constants/alert/alert-severity.constant';

@Component({
  selector: 'app-alert-severity',
  templateUrl: './alert-severity.component.html',
  styleUrls: ['./alert-severity.component.scss']
})
export class AlertSeverityComponent implements OnInit {
  @Input() severity: string | number;
  label: string;
  background: string;

  constructor() {
  }

  ngOnInit() {
    if (typeof this.severity === 'string') {
      this.setBackgroundColor();
    }
    if (typeof this.severity === 'number') {
      this.setBackgroundColorInt();
    }
  }

  private setBackgroundColor() {
    if (HIGH_TEXT === this.severity) {
      this.background = 'border-danger-800 text-danger-800';
      this.label = 'HIGH';
    } else if (MEDIUM_TEXT === this.severity) {
      this.background = 'border-warning-800 text-warning-800';
      this.label = 'MEDIUM';
    } else if (LOW_TEXT === this.severity) {
      this.background = 'border-info-800 text-info-800';
      this.label = 'LOW';
    } else {
      this.background = 'border-slate-800 text-slate-800';
      this.label = 'UNKNOWN';
    }
  }

  private setBackgroundColorInt() {
    if (HIGH === this.severity) {
      this.background = 'border-danger-800 text-danger-800';
      this.label = 'HIGH';
    } else if (MEDIUM === this.severity) {
      this.background = 'border-warning-800 text-warning-800';
      this.label = 'MEDIUM';
    } else if (LOW === this.severity) {
      this.background = 'border-info-800 text-info-800';
      this.label = 'LOW';
    } else {
      this.background = 'border-slate-800 text-slate-800';
      this.label = 'UNKNOWN';
    }
  }

}
