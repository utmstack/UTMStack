import {ChangeDetectionStrategy, Component, Input, OnDestroy, OnInit, Output} from '@angular/core';
import {Observable, Subject} from 'rxjs';
import {concatMap, filter, map, takeUntil, tap} from 'rxjs/operators';
import {TimeWindowsService} from '../../../shared/components/utm-cp-section/time-windows.service';
import { ComplianceReportType } from '../../../shared/type/compliance-report.type';


@Component({
  selector: 'app-compliance-time',
  templateUrl: './compliance-time-windows.component.html',
  styleUrls: ['./compliance-time-windows.component.css'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ComplianceTimeWindowsComponent implements OnInit, OnDestroy {
  @Input() report: ComplianceReportType;
  destroy$ = new Subject<void>();
  time$: Observable<string>;

  constructor(private timeWindowsService: TimeWindowsService) {}

  ngOnInit() {
    this.time$ = this.timeWindowsService.onTimeWindows$
      .pipe(takeUntil(this.destroy$),
            filter(timeWindows => !!timeWindows
                    && timeWindows.reportId === this.report.id),
            map((timeWindows) => timeWindows.time));
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
