import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';
import {LogAnalyzerDataChangeType} from '../type/log-analyzer-data-change.type';

@Injectable({providedIn: 'root'})
export class IndexPatternBehavior {
  $pattern = new BehaviorSubject<LogAnalyzerDataChangeType>(null);
}
