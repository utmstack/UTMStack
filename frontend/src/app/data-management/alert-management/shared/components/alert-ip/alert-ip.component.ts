import {Component, Input, OnInit} from '@angular/core';
import {
  ALERT_DESTINATION_COUNTRY_CODE_FIELD,
  ALERT_DESTINATION_COUNTRY_FIELD,
  ALERT_DESTINATION_IP_FIELD,
  ALERT_SOURCE_COUNTRY_CODE_FIELD,
  ALERT_SOURCE_COUNTRY_FIELD,
  ALERT_SOURCE_IP_FIELD
} from '../../../../../shared/constants/alert/alert-field.constant';
import {getValueFromPropertyPath} from '../../../../../shared/util/get-value-object-from-property-path.util';

@Component({
  selector: 'app-alert-ip',
  templateUrl: './alert-ip.component.html',
  styleUrls: ['./alert-ip.component.scss']
})
export class AlertIpComponent implements OnInit {
  @Input() alert: any;
  @Input() type: 'source' | 'destination';
  ip: string;
  countryCode: string;
  country: string;

  constructor() {
  }

  ngOnInit() {
    if (this.type === 'source') {
      this.ip = getValueFromPropertyPath(this.alert, ALERT_SOURCE_IP_FIELD, null);
      this.country = getValueFromPropertyPath(this.alert, ALERT_SOURCE_COUNTRY_FIELD, null);
      this.countryCode = getValueFromPropertyPath(this.alert, ALERT_SOURCE_COUNTRY_CODE_FIELD, null);
      this.countryCode = this.countryCode ? this.countryCode.toLowerCase() : null;
    } else {
      this.ip = getValueFromPropertyPath(this.alert, ALERT_DESTINATION_IP_FIELD, null);
      this.country = getValueFromPropertyPath(this.alert, ALERT_DESTINATION_COUNTRY_FIELD, null);
      this.countryCode = getValueFromPropertyPath(this.alert, ALERT_DESTINATION_COUNTRY_CODE_FIELD, null);
      this.countryCode = this.countryCode ? this.countryCode.toLowerCase() : null;
    }
    this.ip = this.ip ? this.ip : '-';
  }

}
