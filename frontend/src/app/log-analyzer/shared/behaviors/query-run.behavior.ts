import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

@Injectable({providedIn: 'root'})
export class QueryRunBehavior {
  /**
   * Control status of query run, link between log view and header
   */
  $runQueryFinished = new BehaviorSubject<boolean>(false);
}
