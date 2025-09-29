import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';

@Component({
  selector: 'app-alert-child-column',
  templateUrl: './alert-child-column.component.html',
  styleUrls: ['./alert-child-column.component.scss']
})
export class AlertChildColumnComponent  {

  @Input() alert: any;
  @Input() loadingChildren = false;

  @Output() toggleExpand = new EventEmitter<any>();

  onToggleExpand() {
    this.toggleExpand.emit(this.alert);
  }

}
