import {
  ChangeDetectionStrategy,
  ChangeDetectorRef,
  Component,
  EventEmitter,
  Input, OnChanges,
  OnInit,
  Output
} from '@angular/core';

@Component({
  selector: 'app-utm-selectable-list',
  templateUrl: './selectable-list.component.html',
  styleUrls: ['./selectable-list.component.css']
})
export class SelectableListComponent implements OnChanges {
  @Input() items: {id: any, name: string; selected: boolean }[] = [];
  @Input() label = '';
  @Input() placeholder = 'Search...';
  order: 'asc' | 'desc' = 'asc';
  @Output() itemSelected: EventEmitter<any> = new EventEmitter();
  searchTerm = '';

  constructor() { }

  ngOnChanges(): void {
    console.log('changes:', this.items);
    this.setOrder(this.order);
  }

  onItemSelect(item: any): void {
    this.itemSelected.emit(item);
  }

  setOrder(order: 'asc' | 'desc') {
    this.order = order;

    this.items = [...this.items].sort((a, b) => {
      if (order === 'asc') {
        return a.name.localeCompare(b.name);
      } else {
        return b.name.localeCompare(a.name);
      }
    });
  }
}
