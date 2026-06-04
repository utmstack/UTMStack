import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';

@Component({
  selector: 'app-alert-actions-content',
  templateUrl: './alert-actions-content.component.html',
  styleUrls: ['./alert-actions-content.component.scss']
})
export class AlertActionsContentComponent  {

  @Input() alert: any;
  @Input() dataType: any;
  @Input() tags: any[];
  @Input() loadingChildren = false;
  @Input() alertSelected = [];

  @Output() toggleExpand = new EventEmitter<any>();
  @Output() select = new EventEmitter<any>();
  @Output() filterRow = new EventEmitter<any>();
  @Output() incident = new EventEmitter<any>();
  @Output() openAutomation = new EventEmitter<any>();
  @Output() applyNote = new EventEmitter<any>();
  @Output() applyTags = new EventEmitter<any>();

  onToggleExpand() {
    this.toggleExpand.emit(this.alert);
  }

  onSelect() {
    this.select.emit(this.alert);
  }

  onFilterRow() {
    this.filterRow.emit(this.alert);
  }

  onIncident(event: any) {
    this.incident.emit(event);
  }

  onOpenAutomation() {
    this.openAutomation.emit(this.alert);
  }

  onApplyNote(event: any) {
    this.applyNote.emit(event);
  }

  onApplyTags() {
    this.applyTags.emit();
  }

  isSelected(alert: any): boolean {
    return this.alertSelected.findIndex(value => value.id === alert.id) !== -1;
  }

}
