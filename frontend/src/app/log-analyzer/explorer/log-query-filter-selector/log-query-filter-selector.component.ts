import {Component, OnDestroy, OnInit, ViewChild} from '@angular/core';
import {NgbPopover} from '@ng-bootstrap/ng-bootstrap';
import {Subject} from 'rxjs';

@Component({
  selector: 'app-log-query-filter-selector',
  templateUrl: './log-query-filter-selector.component.html',
  styleUrls: ['./log-query-filter-selector.component.scss']
})
export class LogQueryFilterSelectorComponent implements OnInit, OnDestroy {

  @ViewChild('popoverFilter') popFilter: NgbPopover;
  destroy$ = new Subject<void>();

  constructor() {
  }

  ngOnInit() {

  }



  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  addFilter() {

  }

  openQueryMenu() {

  }
}
