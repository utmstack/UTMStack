import {ChangeDetectionStrategy, Component, Input} from '@angular/core';

@Component({
  selector: 'app-alert-badge-field',
  templateUrl: './alert-badge-field.component.html',
  styleUrls: ['./alert-badge-field.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class AlertBadgeFieldComponent {
  @Input() tooltip: string;
  @Input() value: string;
  @Input() iconClass: string;

  constructor() {
  }
}
