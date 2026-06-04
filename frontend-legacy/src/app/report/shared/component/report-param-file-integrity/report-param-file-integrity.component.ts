import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {ActiveDirectoryService} from '../../../../active-directory/shared/services/active-directory.service';
import {ElasticFilterDefaultTime} from '../../../../shared/components/utm/filters/elastic-filter-time/elastic-filter-time.component';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';

@Component({
  selector: 'app-report-param-file-integrity',
  templateUrl: './report-param-file-integrity.component.html',
  styleUrls: ['./report-param-file-integrity.component.css']
})
export class ReportParamFileIntegrityComponent implements OnInit {
  @Input() defaultTime: ElasticFilterDefaultTime = new ElasticFilterDefaultTime('now-7d', 'now');
  @Input() formatInstant = true;
  @Output() filterChange = new EventEmitter<{ from: string, to: string, top: number }>();
  params = {from: '', to: '', top: 100};

  constructor(private activeDirectoryService: ActiveDirectoryService) {
  }

  ngOnInit() {
  }

  onTimeFilterChange($event: TimeFilterType) {
    this.params.from = $event.timeFrom;
    this.params.to = $event.timeTo;
    this.filterChange.emit(this.params);
  }


}
