import {Component, Input, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmDateFormatEnum} from '../../../shared/enums/utm-date-format.enum';
import {UtmThreadDetailComponent} from './utm-thread-detail/utm-thread-detail.component';

@Component({
  selector: 'app-utm-thread-metrics',
  templateUrl: './utm-thread-metrics.component.html',
  styleUrls: ['./utm-thread-metrics.component.scss']
})
export class UtmThreadMetricsComponent implements OnInit {

  /**
   * object containing thread related metrics
   */
  @Input() threadData: any;
  threadStats: {
    threadDumpAll: number;
    threadDumpRunnable: number;
    threadDumpTimedWaiting: number;
    threadDumpWaiting: number;
    threadDumpBlocked: number;
  };


  constructor(private modalService: NgbModal) {
  }

  ngOnInit() {
    this.threadStats = {
      threadDumpRunnable: 0,
      threadDumpWaiting: 0,
      threadDumpTimedWaiting: 0,
      threadDumpBlocked: 0,
      threadDumpAll: 0
    };

    if (this.threadData !== {}) {
      this.threadData.forEach(value => {
        if (value.threadState === 'RUNNABLE') {
          this.threadStats.threadDumpRunnable += 1;
        } else if (value.threadState === 'WAITING') {
          this.threadStats.threadDumpWaiting += 1;
        } else if (value.threadState === 'TIMED_WAITING') {
          this.threadStats.threadDumpTimedWaiting += 1;
        } else if (value.threadState === 'BLOCKED') {
          this.threadStats.threadDumpBlocked += 1;
        }
      });
    }

    this.threadStats.threadDumpAll =
      this.threadStats.threadDumpRunnable +
      this.threadStats.threadDumpWaiting +
      this.threadStats.threadDumpTimedWaiting +
      this.threadStats.threadDumpBlocked;
  }

  open() {
    const modalRef = this.modalService.open(UtmThreadDetailComponent);
    modalRef.componentInstance.threadDump = this.threadData;
  }

  protected readonly UtmDateFormatEnum = UtmDateFormatEnum;
}
