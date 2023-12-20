import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

@Injectable({
  providedIn: 'root'
})
/**
 * Use this class to emit and subscribe event on tag is added to alert
 */
export class AlertUpdateTagBehavior {
  /**
   * Use this to refresh tag filter list
   */
  $tagRefresh = new BehaviorSubject<boolean>(false);
  $updateTagForAlert = new BehaviorSubject<string>(null);

}
