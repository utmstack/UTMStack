import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {DashboardFilterType} from '../../../../types/filter/dashboard-filter.type';


export enum DashboardFilterLayout {
  ScheduleReportCompliance,
  Default
}
@Component({
  selector: 'app-dashboard-filter-view',
  templateUrl: './dashboard-filter-view.component.html',
  styleUrls: ['./dashboard-filter-view.component.css']
})
export class DashboardFilterViewComponent implements OnInit {
  @Input() filters: DashboardFilterType[];
  @Input() building = false;
  @Input() layout: DashboardFilterLayout = DashboardFilterLayout.Default;
  @Input() position: 'left' | 'right' = 'right';
  @Output() filterAction = new EventEmitter<{ action: 'edit' | 'delete', filter: DashboardFilterType }>();
  DashboardFilterLayout = DashboardFilterLayout;


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
