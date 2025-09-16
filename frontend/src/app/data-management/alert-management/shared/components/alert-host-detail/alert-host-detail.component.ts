import {ChangeDetectionStrategy, Component, Input, OnInit} from '@angular/core';
import {ALERT_FIELDS,
} from '../../../../../shared/constants/alert/alert-field.constant';
import {IncidentOriginTypeEnum} from '../../../../../shared/enums/incident-response/incident-origin-type.enum';
import {UtmAlertType} from '../../../../../shared/types/alert/utm-alert.type';
import {UtmFieldType} from '../../../../../shared/types/table/utm-field.type';
import {AlertFieldService} from '../../services/alert-field.service';
import {isEmpty} from "rxjs/operators";

@Component({
  selector: 'app-alert-host-detail',
  templateUrl: './alert-host-detail.component.html',
  styleUrls: ['./alert-host-detail.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class AlertHostDetailComponent implements OnInit {
  @Input() alert: UtmAlertType;
  @Input() type: 'target' | 'adversary';
  @Input() hideEmptyField = false;
  module = IncidentOriginTypeEnum.ALERT;
  fields: { id: string; fieldType: UtmFieldType }[] = [];

  constructor(private alertFieldService: AlertFieldService) {}

  ngOnInit() {
    this.fields = this.getFields(this.alert[this.type]);
  }

  getFieldByName(name: string): UtmFieldType {
    return this.alertFieldService.findField(ALERT_FIELDS, name);
  }

  isEmpty(): boolean {
    return !this.alert || !this.alert[this.type] || Object.keys(this.alert[this.type]).length === 0;
  }

  getFields(obj: any, prefix = ''): { id: string; fieldType: UtmFieldType }[] {
    let fields: { id: string, fieldType: UtmFieldType}[] = [];

    for (const key in obj) {
      if (obj.hasOwnProperty(key)) {
        const newKey = prefix ? `${prefix}.${key}` : key;

        if (typeof obj[key] === 'object' && obj[key] !== null) {
          fields = [...fields, ...this.getFields(obj[key], newKey)];
        } else {
          const fieldType = this.getFieldByName(`${this.type}.${newKey}`);
          if (fieldType) {
            fields.push({
              id: key,
              fieldType
            });
          }
        }
      }
    }
    return fields;
  }
}
