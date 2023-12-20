import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class AlertTagChangeBehavior {
  /**
   * Use when need refresh alert status count, useful when change alert status
   */
  $updateTag = new BehaviorSubject<string>(null);
}
