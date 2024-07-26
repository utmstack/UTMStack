import {ChangeDetectionStrategy, Component, Input} from '@angular/core';
import { Rule } from 'src/app/rule-management/models/rule.model';
import {IncidentSeverityEnum} from '../../../../../../shared/enums/incident/incident-severity.enum';
import {UtmFieldType} from '../../../../../../shared/types/table/utm-field.type';
import {
  RULE_AVAILABILITY,
  RULE_CONFIDENTIALITY,
  RULE_DATA_TYPES,
  RULE_INTEGRITY
} from '../../../../../models/rule.constant';

@Component({
    selector: 'app-rule-field',
    templateUrl: './rule-field.component.html',
    styleUrls: ['./rule-field.component.css'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class RuleFieldComponent {
    @Input() rule!: Rule;
    @Input() column!: UtmFieldType;

    RULE_DATA_TYPES = RULE_DATA_TYPES;
    RULE_INTEGRITY = RULE_INTEGRITY;
    RULE_AVAILABILITY = RULE_AVAILABILITY;
    RULE_CONFIDENTIALITY = RULE_CONFIDENTIALITY;

    constructor() {}

  getSeverity(value: any): IncidentSeverityEnum {
    switch (value) {
      case 1: return IncidentSeverityEnum.LOW;
      case 2: return IncidentSeverityEnum.MEDIUM;
      case 3: return IncidentSeverityEnum.HIGH;
    }
  }

  isImpactType(field: string): boolean {
    return field === RULE_INTEGRITY || field === RULE_CONFIDENTIALITY || field === RULE_AVAILABILITY;
  }
}
