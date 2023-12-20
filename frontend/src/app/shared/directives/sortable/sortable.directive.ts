import {Directive, EventEmitter, Input, OnChanges, Output, SimpleChanges} from '@angular/core';
import {SORT_ROTATE} from './const/sort.const';
import {SortDirection} from './type/sort-direction.type';
import {SortEvent} from './type/sort-event';

@Directive({
  selector: 'th[appColumnSortable]',
  host: {
    '[class.utm-table-sort-desc]': 'direction=== "asc"',
    '[class.utm-table-sort-asc]': 'direction=== "desc"',
    '[class.utm-table-sort-none]': 'direction === ""',
    '(click)': 'rotate()'
  }
})

export class SortableDirective implements OnChanges {
  @Input() sortable: string;
  @Input() direction: SortDirection = '';
  @Input() isSortable = true;
  @Output() sort = new EventEmitter<SortEvent | string>();
  @Input() sortEvent: SortEvent;

  constructor() {
  }

  rotate() {
    if (this.isSortable) {
      this.direction = SORT_ROTATE[this.direction];
      this.sort.emit({column: this.sortable, direction: this.direction});
    }
  }

  ngOnChanges(changes: SimpleChanges): void {
    // console.log(changes);
    // if (changes && !changes.sortEvent.firstChange) {
    //   if (this.sortEvent.column === this.sortable) {
    //     this.direction = changes.sortEvent.currentValue.direction;
    //     console.log(' change sort event ' + this.direction);
    //   } else {
    //     this.direction = '';
    //   }
    // }
  }

}
