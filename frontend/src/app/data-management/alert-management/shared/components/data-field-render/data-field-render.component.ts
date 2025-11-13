import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {
  ALERT_ADVERSARY_FIELD, ALERT_ADVERSARY_IP_FIELD,
  ALERT_DESTINATION_IP_FIELD, ALERT_ECHOES_FIELD,
  ALERT_IMPACT_FIELD,
  ALERT_NOTE_FIELD,
  ALERT_SEVERITY_FIELD_LABEL,
  ALERT_SOURCE_IP_FIELD,
  ALERT_STATUS_FIELD,
  ALERT_TAGS_FIELD,
  ALERT_TARGET_FIELD,
  ALERT_TARGET_IP_FIELD
} from '../../../../../shared/constants/alert/alert-field.constant';
import {UtmDateFormatEnum} from '../../../../../shared/enums/utm-date-format.enum';
import {UtmAlertType} from '../../../../../shared/types/alert/utm-alert.type';
import {UtmFieldType} from '../../../../../shared/types/table/utm-field.type';
import {extractValueFromObjectByPath} from '../../../../../shared/util/get-value-object-from-property-path.util';
import {EventDataTypeEnum} from '../../enums/event-data-type.enum';

@Component({
  selector: 'app-data-field-render',
  templateUrl: './data-field-render.component.html',
  styleUrls: ['./data-field-render.component.scss']
})
export class DataFieldRenderComponent implements OnInit {
  @Input() data: UtmAlertType;
  @Input() field: UtmFieldType;
  @Input() showStatusChange: boolean;
  @Input() dataType: EventDataTypeEnum;
  @Input() tags: any[];
  @Input() isEcho = false;
  @Output() refreshData = new EventEmitter<boolean>();
  STATUS_FIELD = ALERT_STATUS_FIELD;
  SEVERITY_LABEL_FIELD = ALERT_SEVERITY_FIELD_LABEL;
  NOTE_FIELD = ALERT_NOTE_FIELD;
  TAGS_FIELD = ALERT_TAGS_FIELD;
  ALERT_IMPACT_FIELD = ALERT_IMPACT_FIELD;
  DESTINATION_IP_FIELD = ALERT_DESTINATION_IP_FIELD;
  ALERT_TARGET_FIELD = ALERT_TARGET_FIELD;
  ALERT_ADVERSARY_FIELD = ALERT_ADVERSARY_FIELD;
  utmFormatDate = UtmDateFormatEnum.UTM_SHORT_UTC;
  ALERT_TARGET_IP_FIELD = ALERT_TARGET_IP_FIELD;
  ALERT_ADVERSARY_IP_FIELD = ALERT_ADVERSARY_IP_FIELD;
  ALERT_ECHOES_FIELD = ALERT_ECHOES_FIELD;

  constructor() {
  }

  ngOnInit() {
  }

  resolveValue(alert: any, td: UtmFieldType) {
    return extractValueFromObjectByPath(alert, td);
  }

  resolveSeverity(alert: any) {
    return this.data.severityLabel;
  }

  onStatusChange($event: number) {
    this.refreshData.emit(true);
  }
}
