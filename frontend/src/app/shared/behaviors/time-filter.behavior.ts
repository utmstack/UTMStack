import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

@Injectable({providedIn: 'root'})
export class TimeFilterBehavior {
  $time = new BehaviorSubject<{ from: any, to: any }>(null);
}
