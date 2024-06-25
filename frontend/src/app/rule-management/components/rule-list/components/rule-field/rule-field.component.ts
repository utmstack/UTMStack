import {ChangeDetectionStrategy, Component, Input} from '@angular/core';
import { Rule } from 'src/app/rule-management/models/rule.model';
import {UtmFieldType} from '../../../../../shared/types/table/utm-field.type';
import {RULE_DATA_TYPES, RULE_REFERENCES} from '../../../../models/rule.constant';

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
    RULE_REFERENCES = RULE_REFERENCES;

    constructor() {}
}
