import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class AlertStatusBehavior {
  /**
   * Use when need refresh alert status count, useful when change alert status
   */
  $updateStatus = new BehaviorSubject<boolean>(null);
}
