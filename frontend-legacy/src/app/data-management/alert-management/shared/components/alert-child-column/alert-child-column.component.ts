import {Component, EventEmitter, Input, Output} from '@angular/core';
import {UtmAlertType} from 'src/app/shared/types/alert/utm-alert.type';

@Component({
  selector: 'app-alert-child-column',
  templateUrl: './alert-child-column.component.html',
  styleUrls: ['./alert-child-column.component.scss']
})
export class AlertChildColumnComponent  {

  @Input() alert: UtmAlertType;
  @Input() loadingChildren = false;

  @Output() toggleExpand = new EventEmitter<any>();

  onToggleExpand(): void {
    if (this.alert.hasChildren) {
        this.alert.expanded = !this.alert.expanded;
        this.toggleExpand.emit(this.alert);
    }
  }

}
