import {Component, Input, OnInit} from '@angular/core';
import {
  ALERT_DESTINATION_ASN_FIELD,
  ALERT_DESTINATION_ASO_FIELD,
  ALERT_DESTINATION_CITY_FIELD,
  ALERT_DESTINATION_COUNTRY_FIELD,
  ALERT_DESTINATION_HOSTNAME_FIELD,
  ALERT_DESTINATION_IP_FIELD,
  ALERT_DESTINATION_IS_AN_PROXY_FIELD,
  ALERT_DESTINATION_IS_SATELLITE_FIELD,
  ALERT_DESTINATION_PORT_FIELD,
  ALERT_DESTINATION_USER_FIELD,
  ALERT_FIELDS,
  ALERT_SOURCE_ASN_FIELD,
  ALERT_SOURCE_ASO_FIELD,
  ALERT_SOURCE_CITY_FIELD,
  ALERT_SOURCE_COUNTRY_FIELD,
  ALERT_SOURCE_HOSTNAME_FIELD,
  ALERT_SOURCE_IP_FIELD,
  ALERT_SOURCE_IS_AN_PROXY_FIELD,
  ALERT_SOURCE_IS_SATELLITE_FIELD,
  ALERT_SOURCE_PORT_FIELD,
  ALERT_SOURCE_USER_FIELD
} from '../../../../../shared/constants/alert/alert-field.constant';
import {IncidentOriginTypeEnum} from '../../../../../shared/enums/incident-response/incident-origin-type.enum';
import {UtmAlertType} from '../../../../../shared/types/alert/utm-alert.type';
import {UtmFieldType} from '../../../../../shared/types/table/utm-field.type';

@Component({
  selector: 'app-alert-host-detail',
  templateUrl: './alert-host-detail.component.html',
  styleUrls: ['./alert-host-detail.component.scss']
})
export class AlertHostDetailComponent implements OnInit {
  @Input() alert: UtmAlertType;
  @Input() type: 'source' | 'destination';
  @Input() hideEmptyField = false;

  SOURCE_HOSTNAME_FIELD = ALERT_SOURCE_HOSTNAME_FIELD;
  SOURCE_IP_FIELD = ALERT_SOURCE_IP_FIELD;
  SOURCE_PORT_FIELD = ALERT_SOURCE_PORT_FIELD;
  SOURCE_COUNTRY = ALERT_SOURCE_COUNTRY_FIELD;
  SOURCE_CITY = ALERT_SOURCE_CITY_FIELD;
  SOURCE_ASN = ALERT_SOURCE_ASN_FIELD;
  SOURCE_ASO = ALERT_SOURCE_ASO_FIELD;
  SOURCE_IS_SATELLITE = ALERT_SOURCE_IS_SATELLITE_FIELD;
  SOURCE_IS_AN_PROXY = ALERT_SOURCE_IS_AN_PROXY_FIELD;
  SOURCE_USER_FIELD = ALERT_SOURCE_USER_FIELD;

  DESTINATION_HOSTNAME_FIELD = ALERT_DESTINATION_HOSTNAME_FIELD;
  DESTINATION_IP_FIELD = ALERT_DESTINATION_IP_FIELD;
  DESTINATION_PORT_FIELD = ALERT_DESTINATION_PORT_FIELD;
  DESTINATION_COUNTRY = ALERT_DESTINATION_COUNTRY_FIELD;
  DESTINATION_CITY = ALERT_DESTINATION_CITY_FIELD;
  DESTINATION_ASN = ALERT_DESTINATION_ASN_FIELD;
  DESTINATION_ASO = ALERT_DESTINATION_ASO_FIELD;
  DESTINATION_IS_SATELLITE = ALERT_DESTINATION_IS_SATELLITE_FIELD;
  DESTINATION_IS_AN_PROXY = ALERT_DESTINATION_IS_AN_PROXY_FIELD;
  DESTINATION_USER_FIELD = ALERT_DESTINATION_USER_FIELD;

  module = IncidentOriginTypeEnum.ALERT;

  constructor() {
  }

  ngOnInit() {
  }

  getFieldByName(name): UtmFieldType {
    return ALERT_FIELDS.find(value => value.field === name);
  }

  getHost(): string {
    return this.type === 'source' ? this.alert.source.ip : this.alert.destination.ip;
  }

  getHostName(): string {
    return this.type === 'source' ? this.alert.source.host : this.alert.destination.host;
  }

  getAlertId() {
    return this.alert.id;
  }
}
