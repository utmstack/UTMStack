import { Component, EventEmitter, Input, Output } from '@angular/core';
import {UtmComplianceQueryConfigType} from '../../../type/compliance-query-config.type';

@Component({
  selector: 'app-utm-compliance-query-list',
  templateUrl: './utm-compliance-query-list.component.html',
  styleUrls: ['./utm-compliance-query-list.component.scss']
})
export class UtmComplianceQueryListComponent {

  @Input() queries: UtmComplianceQueryConfigType[] = [];

  @Output() add = new EventEmitter<UtmComplianceQueryConfigType>();
  @Output() edit = new EventEmitter<{ index: number, query: UtmComplianceQueryConfigType }>();
  @Output() delete = new EventEmitter<number>();

  editingIndex: number = null;

  startEdit(index: number) {
    this.editingIndex = index;
  }

  finishEdit(query: UtmComplianceQueryConfigType) {
    this.edit.emit({ index: this.editingIndex, query });
    this.editingIndex = null;
  }

  cancelEdit() {
    this.editingIndex = null;
  }

  remove(index: number) {
    this.delete.emit(index);
  }
}
