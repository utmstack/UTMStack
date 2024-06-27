import {ChangeDetectionStrategy, Component, Input} from '@angular/core';
import { Rule } from 'src/app/rule-management/models/rule.model';

@Component({
    selector: 'app-rule-detail',
    templateUrl: './rule-detail.component.html',
    styleUrls: ['./rule-detail.component.css'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class RuleDetailComponent {
    @Input() rule!: Rule;

    constructor() {}
}
