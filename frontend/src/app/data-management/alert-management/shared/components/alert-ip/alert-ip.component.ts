import {Component, Input, OnChanges, OnInit, SimpleChanges} from '@angular/core';
import {
  ALERT_ADVERSARY_GEOLOCATION_COUNTRY_CODE_FIELD,
  ALERT_ADVERSARY_GEOLOCATION_COUNTRY_FIELD,
  ALERT_ADVERSARY_IP_FIELD,
  ALERT_DESTINATION_COUNTRY_CODE_FIELD,
  ALERT_DESTINATION_COUNTRY_FIELD,
  ALERT_DESTINATION_IP_FIELD,
  ALERT_SOURCE_COUNTRY_CODE_FIELD,
  ALERT_SOURCE_COUNTRY_FIELD,
  ALERT_SOURCE_IP_FIELD,
  ALERT_TARGET_GEOLOCATION_COUNTRY_CODE_FIELD,
  ALERT_TARGET_GEOLOCATION_COUNTRY_FIELD,
  ALERT_TARGET_IP_FIELD
} from '../../../../../shared/constants/alert/alert-field.constant';
import {getValueFromPropertyPath} from '../../../../../shared/util/get-value-object-from-property-path.util';

@Component({
  selector: 'app-alert-ip',
  templateUrl: './alert-ip.component.html',
  styleUrls: ['./alert-ip.component.scss']
})
export class AlertIpComponent implements OnInit, OnChanges {
  @Input() alert: any;
  @Input() type: 'source' | 'destination' | 'target' | 'adversary';
  ipField: string;
  countryField: string;
  countryCodeField: string;
  ip: string;
  countryCode: string;
  country: string;

  constructor() {
  }

  ngOnChanges(changes: SimpleChanges): void {
    this.loadFields();
  }

  ngOnInit() {
      this.ip = getValueFromPropertyPath(this.alert, this.ipField, null);
      this.country = getValueFromPropertyPath(this.alert, this.countryField, null);
      this.countryCode = getValueFromPropertyPath(this.alert, this.countryCodeField, null);
      this.countryCode = this.countryCode ? this.countryCode.toLowerCase() : null;

      this.ip = this.ip ? this.ip : '-';
  }

  loadFields() {
    switch (this.type) {
      case 'source':
        this.ipField = ALERT_SOURCE_IP_FIELD;
        this.countryField = ALERT_SOURCE_COUNTRY_FIELD;
        this.countryCodeField = ALERT_SOURCE_COUNTRY_CODE_FIELD;
        break;
      case 'destination':
        this.ipField = ALERT_DESTINATION_IP_FIELD;
        this.countryField = ALERT_DESTINATION_COUNTRY_FIELD;
        this.countryCodeField = ALERT_DESTINATION_COUNTRY_CODE_FIELD;
        break;
      case 'target':
        this.ipField = ALERT_TARGET_IP_FIELD;
        this.countryField = ALERT_TARGET_GEOLOCATION_COUNTRY_FIELD;
        this.countryCodeField = ALERT_TARGET_GEOLOCATION_COUNTRY_CODE_FIELD;
        break;
      case 'adversary':
        this.ipField = ALERT_ADVERSARY_IP_FIELD;
        this.countryField = ALERT_ADVERSARY_GEOLOCATION_COUNTRY_FIELD;
        this.countryCodeField = ALERT_ADVERSARY_GEOLOCATION_COUNTRY_CODE_FIELD;
        break;
      default:
        break;
    }
  }
}
