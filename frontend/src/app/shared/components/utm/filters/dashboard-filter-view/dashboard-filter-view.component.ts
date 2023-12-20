import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {DashboardFilterType} from '../../../../types/filter/dashboard-filter.type';

@Component({
  selector: 'app-dashboard-filter-view',
  templateUrl: './dashboard-filter-view.component.html',
  styleUrls: ['./dashboard-filter-view.component.css']
})
export class DashboardFilterViewComponent implements OnInit {
  @Input() filters: DashboardFilterType[];
  @Input() building = false;
  @Input() position: 'left' | 'right' = 'right';
  @Output() filterAction = new EventEmitter<{ action: 'edit' | 'delete', filter: DashboardFilterType }>();


  constructor() {
  }

  ngOnInit() {

  }

  select($event: unknown, filter: DashboardFilterType) {

  }


  doAction(action: 'edit' | 'delete', filter: DashboardFilterType) {
    this.filterAction.emit({action, filter});
  }


}
