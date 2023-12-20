import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';

@Component({
  selector: 'app-active-directory-filter',
  templateUrl: './active-directory-filter.component.html',
  styleUrls: ['./active-directory-filter.component.scss']
})
export class ActiveDirectoryFilterComponent implements OnInit {
  @Output() adFilterChange = new EventEmitter<any>();
  @Input() sAMAccountName: string;
  filter = {};

  constructor() {
  }

  ngOnInit() {
    this.filter['sAMAccountName.equals'] = this.sAMAccountName ? this.sAMAccountName : '';
    if (this.sAMAccountName) {
      this.adFilterChange.emit(this.filter);
    }
  }

  searchByText() {
    this.adFilterChange.emit(this.filter);
  }

  searchByCreation($event: TimeFilterType) {
    this.filter['whenCreated.greaterThanOrEqual'] = $event.timeFrom;
    this.filter['whenCreated.lessThanOrEqual'] = $event.timeTo;
    this.adFilterChange.emit(this.filter);
  }

  searchByLastEvent($event: TimeFilterType) {
    this.filter['realLastLogon.greaterThanOrEqual'] = $event.timeFrom;
    this.filter['realLastLogon.lessThanOrEqual'] = $event.timeTo;
    this.adFilterChange.emit(this.filter);
  }

}
