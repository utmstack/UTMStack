import {Component, OnInit} from '@angular/core';
import {LocalStorageService} from 'ngx-webstorage';
import {RefreshService} from '../../../../services/util/refresh.service';

export const TIME_REFRESH = 'TIME_REFRESH';
@Component({
  selector: 'app-refresh-filter',
  templateUrl: 'refresh-filter.component.html',
  styleUrls: ['refresh-filter.component.css']
})
export class RefreshFilterComponent implements OnInit {
  show = false;
  refreshIntervals = [
    { label: '1 min', value: 60000 },
    { label: '5 min', value: 300000 },
    { label: '10 min', value: 300000 },
    { label: 'Pause', value: 0 },
  ];
  selectedInterval = 0;
  constructor(private refreshService: RefreshService,
              private storage: LocalStorageService) {
  }
  ngOnInit(): void {
    const time = this.storage.retrieve(TIME_REFRESH);
    if (time && time > 0) {
      this.selectedInterval = time;
      this.apply();
    }
  }

  open(){
    this.show = true;
  }

  apply() {
    this.storage.store(TIME_REFRESH, this.selectedInterval);
    if (this.selectedInterval !== 0) {
      this.refreshService.setRefreshInterval(this.selectedInterval);
    } else {
      this.refreshService.stopInterval();
    }
  }
}
