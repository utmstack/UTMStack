import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {TimeFilterType} from '../../../../../../shared/types/time-filter.type';

@Component({
  selector: 'app-target-filter',
  templateUrl: './target-filter.component.html',
  styleUrls: ['./target-filter.component.scss']
})
export class TargetFilterComponent implements OnInit {
  @Output() targetFilterChange = new EventEmitter<any>();
  filter = {};

  constructor() {
  }

  ngOnInit() {
  }

  searchByText() {
    this.targetFilterChange.emit(this.filter);
  }

  searchByCreation($event: TimeFilterType) {
    this.filter['created.greaterThan'] = $event.timeFrom.substring(0, $event.timeFrom.lastIndexOf(':'));
    this.filter['created.lessThan'] = $event.timeTo.substring(0, $event.timeFrom.lastIndexOf(':'));
    this.targetFilterChange.emit(this.filter);
  }
}
