import {Component, Input, OnChanges, OnInit, SimpleChanges} from '@angular/core';
import {
  ALERT_ADVERSARY_BYTES_SENT_FIELD,
  ALERT_ADVERSARY_DOMAIN_FIELD, ALERT_ADVERSARY_GEOLOCATION_ASN_FIELD, ALERT_ADVERSARY_GEOLOCATION_ASO_FIELD,
  ALERT_ADVERSARY_GEOLOCATION_LATITUDE_FIELD, ALERT_ADVERSARY_GEOLOCATION_LONGITUDE_FIELD, ALERT_ADVERSARY_HOST_FIELD,
  ALERT_ADVERSARY_IP_FIELD,
  ALERT_ADVERSARY_URL_FIELD, ALERT_ADVERSARY_USER_FIELD,
  ALERT_TARGET_BYTES_SENT_FIELD,
  ALERT_TARGET_DOMAIN_FIELD,
  ALERT_TARGET_FIELD, ALERT_TARGET_GEOLOCATION_ASN_FIELD, ALERT_TARGET_GEOLOCATION_ASO_FIELD,
  ALERT_TARGET_GEOLOCATION_LATITUDE_FIELD,
  ALERT_TARGET_GEOLOCATION_LONGITUDE_FIELD, ALERT_TARGET_HOST_FIELD,
  ALERT_TARGET_IP_FIELD,
  ALERT_TARGET_URL_FIELD, ALERT_TARGET_USER_FIELD
} from '../../../../../shared/constants/alert/alert-field.constant';
import {UtmAlertType} from '../../../../../shared/types/alert/utm-alert.type';
import {UtmFieldType} from '../../../../../shared/types/table/utm-field.type';
import {getValueFromPropertyPath} from '../../../../../shared/util/get-value-object-from-property-path.util';
import {AlertFieldService} from '../../services/alert-field.service';

@Component({
  selector: 'app-alert-entity-display',
  templateUrl: './alert-entity-display.component.html',
  styleUrls: ['./alert-entity-display.component.css']
})
export class AlertEntityDisplayComponent implements OnInit, OnChanges {
  ALERT_ADVERSARY_IP_FIELD = ALERT_ADVERSARY_IP_FIELD;
  ALERT_TARGET_IP_FIELD = ALERT_TARGET_IP_FIELD;
  ALERT_ADVERSARY_BYTES_SENT_FIELD = ALERT_ADVERSARY_BYTES_SENT_FIELD;
  ALERT_TARGET_BYTES_SENT_FIELD = ALERT_TARGET_BYTES_SENT_FIELD;
  ALERT_ADVERSARY_URL_FIELD = ALERT_ADVERSARY_URL_FIELD;
  ALERT_TARGET_URL_FIELD = ALERT_TARGET_URL_FIELD;
  ALERT_TARGET_DOMAIN_FIELD = ALERT_TARGET_DOMAIN_FIELD;
  ALERT_ADVERSARY_DOMAIN_FIELD = ALERT_ADVERSARY_DOMAIN_FIELD;
  ALERT_ADVERSARY_HOST_FIELD = ALERT_ADVERSARY_HOST_FIELD;
  ALERT_ADVERSARY_USER_FIELD = ALERT_ADVERSARY_USER_FIELD;
  ALERT_TARGET_USER_FIELD = ALERT_TARGET_USER_FIELD;
  ALERT_TARGET_HOST_FIELD = ALERT_TARGET_HOST_FIELD;
  ALERT_TARGET_GEOLOCATION_LATITUDE_FIELD = ALERT_TARGET_GEOLOCATION_LATITUDE_FIELD;
  ALERT_ADVERSARY_GEOLOCATION_LATITUDE_FIELD = ALERT_ADVERSARY_GEOLOCATION_LATITUDE_FIELD;
  ALERT_TARGET_GEOLOCATION_LONGITUDE_FIELD = ALERT_TARGET_GEOLOCATION_LONGITUDE_FIELD;
  ALERT_ADVERSARY_GEOLOCATION_LONGITUDE_FIELD = ALERT_ADVERSARY_GEOLOCATION_LONGITUDE_FIELD;
  ALERT_TARGET_GEOLOCATION_ASN_FIELD = ALERT_TARGET_GEOLOCATION_ASN_FIELD;
  ALERT_ADVERSARY_GEOLOCATION_ASN_FIELD = ALERT_ADVERSARY_GEOLOCATION_ASN_FIELD;
  ALERT_TARGET_GEOLOCATION_ASO_FIELD = ALERT_TARGET_GEOLOCATION_ASO_FIELD;
  ALERT_ADVERSARY_GEOLOCATION_ASO_FIELD = ALERT_ADVERSARY_GEOLOCATION_ASO_FIELD;

   @Input() alert: UtmAlertType;
   @Input() key: string;
   @Input() field: UtmFieldType;
   fields = [];
   geolocationFields = [];
   type: 'target' | 'adversary';

   constructor(private alertFieldService: AlertFieldService) { }

  ngOnChanges(changes: SimpleChanges): void {
     this.type = this.key === ALERT_TARGET_FIELD ? 'target' : 'adversary';
  }

  ngOnInit() {
     if (this.alert[this.key]) {
       this.fields = Object.keys(this.alert[this.key]);
       if (this.alert[this.key].geolocation) {
         this.geolocationFields = Object.keys(this.alert[this.key].geolocation);
       }
     }
  }

  getValue(field: string){
    return getValueFromPropertyPath(this.alert, field, null);
  }

  isInclude(searchEl: string, arr: string[]): boolean {
    return arr.includes(searchEl);
  }
}
