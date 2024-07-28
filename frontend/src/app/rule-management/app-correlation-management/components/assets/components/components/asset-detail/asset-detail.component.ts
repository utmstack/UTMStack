import {ChangeDetectionStrategy, Component, Input} from '@angular/core';
import {Asset, Rule} from 'src/app/rule-management/models/rule.model';
import {IncidentSeverityEnum} from '../../../../../../../shared/enums/incident/incident-severity.enum';

@Component({
    selector: 'app-asset-detail',
    templateUrl: './asset-detail.component.html',
    styleUrls: ['./asset-detail.component.css'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class AssetDetailComponent {
    @Input() asset!: Asset;

    constructor() {}

  getSeverity(value: any): IncidentSeverityEnum {
    switch (value) {
      case 1: return IncidentSeverityEnum.LOW;
      case 2: return IncidentSeverityEnum.MEDIUM;
      case 3: return IncidentSeverityEnum.HIGH;

      default: return IncidentSeverityEnum.UNIMPORTANT;
    }
  }
}
