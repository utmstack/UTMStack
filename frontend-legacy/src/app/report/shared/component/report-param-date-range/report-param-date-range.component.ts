import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {ElasticFilterDefaultTime} from '../../../../shared/components/utm/filters/elastic-filter-time/elastic-filter-time.component';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';

@Component({
  selector: 'app-report-param-date-range',
  templateUrl: './report-param-date-range.component.html',
  styleUrls: ['./report-param-date-range.component.css']
})
export class ReportParamDateRangeComponent implements OnInit {
  @Input() defaultTime: ElasticFilterDefaultTime = new ElasticFilterDefaultTime('now-7d', 'now');
  @Input() formatInstant = false;
  @Output() rangeChange = new EventEmitter<{ from: string, to: string }>();

  constructor() {
  }

  ngOnInit() {
  }

  onTimeFilterChange($event: TimeFilterType) {
    this.rangeChange.emit({from: $event.timeFrom, to: $event.timeTo});
  }
}
