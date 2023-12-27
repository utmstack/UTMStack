import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

@Injectable({
  providedIn: 'root'
})
/**
 * Use this class to emit and subscribe event on alert solution change
 */
export class AlertUpdateSolutionBehavior {
  /**
   * Use this to refresh to view progress of alert documentation solution update
   */
  $updateSolution = new BehaviorSubject<'init' | 'error' | 'done'>(null);
}
