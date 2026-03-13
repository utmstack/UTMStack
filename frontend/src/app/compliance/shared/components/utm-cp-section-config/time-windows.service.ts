import {BehaviorSubject} from 'rxjs';
import {Injectable} from '@angular/core';

@Injectable()
export class TimeWindowsService {
  private timeWindowsSubject = new BehaviorSubject<{reportId: number, time: string}>(null);
  onTimeWindows$ = this.timeWindowsSubject.asObservable();

  changeTimeWindows(reportTimeWindows: {reportId: number, time: string}) {
    this.timeWindowsSubject.next(reportTimeWindows);
  }
}
