import {Component, EventEmitter, Input, OnChanges, OnInit, Output, SimpleChanges} from '@angular/core';
import {SortDirection} from '../../../../directives/sortable/type/sort-direction.type';
import {SortEvent} from '../../../../directives/sortable/type/sort-event';
import {SortByType} from '../../../../types/sort-by.type';

@Component({
  selector: 'app-sort-by',
  templateUrl: './sort-by.component.html',
  styleUrls: ['./sort-by.component.scss']
})
export class SortByComponent implements OnInit, OnChanges {
  @Input() fields: SortByType[];
  @Input() sortEvent: SortEvent;
  @Output() sortBy = new EventEmitter<SortEvent>();
  @Input() default: string;
  @Input() fieldDefault: string;
  label = 'Default';
  sort = true;
  field = '_id';


  constructor() {
  }

  ngOnInit() {
    if (this.default) {
      this.label = this.default;
    }
    if (this.fieldDefault) {
      this.field = this.fieldDefault;
    }
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (changes && changes.sortEvent && !changes.sortEvent.firstChange) {
      this.label = this.getFieldLabelByColumn(changes.sortEvent.currentValue.column);
      this.sort = changes.sortEvent.currentValue.direction === 'desc';
    }
  }

  sortByFiled(field: SortByType) {
    this.label = field.fieldName;
    this.field = field.field;
    this.sortBy.emit({column: this.field, direction: this.checkOrder()});
  }

  changeSort() {
    this.sort = this.sort ? false : true;
    this.sortBy.emit({column: this.field, direction: this.checkOrder()});
  }

  getFieldLabelByColumn(col: string): string {
    return this.fields[this.fields.findIndex(value => value.field === col)].fieldName;
  }

  checkOrder(): SortDirection {
    return this.sort ? 'desc' : 'asc';
  }
}
