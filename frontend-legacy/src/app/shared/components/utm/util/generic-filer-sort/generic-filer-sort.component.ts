import {Component, EventEmitter, OnInit, Output} from '@angular/core';

@Component({
  selector: 'app-generic-filer-sort',
  templateUrl: './generic-filer-sort.component.html',
  styleUrls: ['./generic-filer-sort.component.scss']
})
export class GenericFilerSortComponent implements OnInit {
  valuesSorts: AlertGenericFilerSort[] =
    [
      {icon: 'icon-sort-numeric-asc', label: 'Count asc', orderByCount: true, sortAsc: true},
      {icon: 'icon-sort-numberic-desc', label: 'Count desc', orderByCount: true, sortAsc: false},
      {icon: 'icon-sort-alpha-asc', label: 'Alphabetical asc', orderByCount: false, sortAsc: true},
      {icon: 'icon-sort-alpha-desc', label: 'Alphabetical desc', orderByCount: false, sortAsc: false},
    ];

  sortSelected: AlertGenericFilerSort = this.valuesSorts[1];
  @Output() sortChange = new EventEmitter<{ orderByCount: boolean, sortAsc: boolean }>();

  constructor() {
  }

  ngOnInit() {
  }

  sortBy(sort: AlertGenericFilerSort) {
    this.sortSelected = sort;
    this.sortChange.emit({orderByCount: sort.orderByCount, sortAsc: sort.sortAsc});
  }
}

export class AlertGenericFilerSort {
  icon: string;
  label: string;
  orderByCount: boolean;
  sortAsc: boolean;
}
