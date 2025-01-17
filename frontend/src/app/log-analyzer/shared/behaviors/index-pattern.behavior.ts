import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';
import {LogAnalyzerDataChangeType} from '../type/log-analyzer-data-change.type';

@Injectable({providedIn: 'root'})
export class IndexPatternBehavior {
  private patternBehavior = new BehaviorSubject<LogAnalyzerDataChangeType>(null);
  pattern$ = this.patternBehavior.asObservable();

  constructor() {
  }

  changePattern(pattern: LogAnalyzerDataChangeType) {
    this.patternBehavior.next(pattern);
  }
}
