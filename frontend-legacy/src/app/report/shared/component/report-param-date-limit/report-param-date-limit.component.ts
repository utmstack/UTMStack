import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {ElasticFilterDefaultTime} from '../../../../shared/components/utm/filters/elastic-filter-time/elastic-filter-time.component';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';

@Component({
  selector: 'app-report-param-date-limit',
  templateUrl: './report-param-date-limit.component.html',
  styleUrls: ['./report-param-date-limit.component.css']
})
export class ReportParamDateLimitComponent implements OnInit {
  @Input() defaultTime: ElasticFilterDefaultTime = new ElasticFilterDefaultTime('now-7d', 'now');
  @Input() formatInstant = true;
  @Output() filterChange = new EventEmitter<{ from: string, to: string, top: number }>();
  params = {from: '', to: '', top: 100};

  constructor() {
  }

  ngOnInit() {
  }

  onTimeFilterChange($event: TimeFilterType) {
    this.params.from = $event.timeFrom;
    this.params.to = $event.timeTo;
    this.filterChange.emit(this.params);
  }

  onLimitChange($event: number) {
    this.params.top = $event;
    this.filterChange.emit(this.params);
  }
}
