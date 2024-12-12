import {ChangeDetectionStrategy, Component, Input, OnChanges, OnInit, SimpleChanges} from '@angular/core';
import {
  ALERT_ADVERSARY_FIELD,
  ALERT_TARGET_FIELD
} from '../../../../../shared/constants/alert/alert-field.constant';
import {UtmAlertType} from '../../../../../shared/types/alert/utm-alert.type';
import * as console from "console";

@Component({
  selector: 'app-alert-entity-display',
  templateUrl: './alert-entity-display.component.html',
  styleUrls: ['./alert-entity-display.component.css'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class AlertEntityDisplayComponent implements OnInit, OnChanges {
  ALERT_TARGET_FIELD = ALERT_TARGET_FIELD;
  ALERT_ADVERSARY_FIELD = ALERT_ADVERSARY_FIELD;
   @Input() alert: UtmAlertType;
   @Input() key: string;
   fields = [];
   geolocationFields = [];
   type: 'target' | 'adversary';

   constructor() { }

  ngOnChanges(changes: SimpleChanges): void {
     this.type = this.key === ALERT_TARGET_FIELD ? 'target' : 'adversary';
  }

  ngOnInit() {
    this.fields = Object.keys(this.alert[this.key]);
    this.geolocationFields = Object.keys(this.alert[this.key].geolocation);
  }

}
