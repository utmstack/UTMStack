import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {TimeFilterType} from '../../../../../../shared/types/time-filter.type';

@Component({
  selector: 'app-port-filter',
  templateUrl: './port-filter.component.html',
  styleUrls: ['./port-filter.component.scss']
})
export class PortFilterComponent implements OnInit {
  @Output() portFilterChange = new EventEmitter<any>();
  filter = {};

  constructor() {
  }

  ngOnInit() {
  }

  searchByUdp() {
    if (this.filter && this.filter['udp.greaterThan'] <= this.filter['udp.lessThan']) {
      this.portFilterChange.emit(this.filter);
    }
  }

  searchByTcp() {
    if (this.filter && this.filter['tcp.greaterThan'] <= this.filter['tcp.lessThan']) {
      this.portFilterChange.emit(this.filter);
    }
  }

  searchByName() {
    this.portFilterChange.emit(this.filter);
  }

  changeFilterTime($event: TimeFilterType) {
    this.filter['created.greaterThan'] = $event.timeFrom.substring(0, $event.timeFrom.lastIndexOf(':'));
    this.filter['created.lessThan'] = $event.timeTo.substring(0, $event.timeFrom.lastIndexOf(':'));
    this.portFilterChange.emit(this.filter);
  }

}
