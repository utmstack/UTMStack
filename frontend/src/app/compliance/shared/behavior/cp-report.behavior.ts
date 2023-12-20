import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

@Injectable({providedIn: 'root'})
export class CpReportBehavior {
  $reportUpdate: BehaviorSubject<string> = new BehaviorSubject<string>(null);
}
