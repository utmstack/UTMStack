import {Component, Input, OnInit} from '@angular/core';
import {ALERT_FULL_LOG_FIELD} from '../../../../../shared/constants/alert/alert-field.constant';
import {UtmAlertType} from '../../../../../shared/types/alert/utm-alert.type';
import {getValueFromPropertyPath} from '../../../../../shared/util/get-value-object-from-property-path.util';

@Component({
  selector: 'app-alert-full-log',
  templateUrl: './alert-full-log.component.html',
  styleUrls: ['./alert-full-log.component.scss']
})
export class AlertFullLogComponent implements OnInit {
  @Input() alert: UtmAlertType;
  fullLog: object;

  constructor() {
  }

  ngOnInit() {
    this.fullLog = getValueFromPropertyPath(this.alert, ALERT_FULL_LOG_FIELD, null);
  }
}
