import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class AlertUpdateHistoryBehavior {
  /**
   * Use tp have to refresh alert history list
   */
  $refreshHistory = new BehaviorSubject<boolean>(null);

}
