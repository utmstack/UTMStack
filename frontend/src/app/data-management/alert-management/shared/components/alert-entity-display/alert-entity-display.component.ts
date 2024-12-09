import {ChangeDetectionStrategy, Component, Input, OnInit} from '@angular/core';
import {
  ALERT_DESTINATION_IP_FIELD, ALERT_SOURCE_FIELD,
  ALERT_SOURCE_IP_FIELD
} from '../../../../../shared/constants/alert/alert-field.constant';
import {UtmAlertType} from '../../../../../shared/types/alert/utm-alert.type';

@Component({
  selector: 'app-alert-entity-display',
  templateUrl: './alert-entity-display.component.html',
  styleUrls: ['./alert-entity-display.component.css',],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class AlertEntityDisplayComponent implements OnInit {
   readonly DESTINATION_IP_FIELD = ALERT_DESTINATION_IP_FIELD;
   readonly SOURCE_IP_FIELD = ALERT_SOURCE_IP_FIELD;
   @Input() alert: UtmAlertType;
   @Input() key: string;
   fields = [];
  constructor() { }

  ngOnInit() {
    console.log('key', this.key);
    this.fields = Object.keys(this.alert[this.key]);
  }

  protected readonly ALERT_SOURCE_FIELD = ALERT_SOURCE_FIELD;
}
