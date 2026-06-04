import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {DATA_LIMIT_RANGE} from '../../../../constants/compliance.constants';

@Component({
  selector: 'app-utm-data-limit',
  templateUrl: './utm-data-limit.component.html',
  styleUrls: ['./utm-data-limit.component.scss']
})
export class UtmDataLimitComponent implements OnInit {
  limitRange = DATA_LIMIT_RANGE;
  limit = 100;
  @Output() limitChange = new EventEmitter<number>();

  constructor() {
  }

  ngOnInit() {
    this.limitChange.emit(this.limit);
  }

  applyLimit(lim: number) {
    this.limit = lim;
    this.limitChange.emit(lim);
  }
}
