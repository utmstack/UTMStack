import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

@Injectable({providedIn: 'root'})
export class IndexFieldController {
  /**
   * Use BehaviorSubject to easily add and remove field on log-analyzer-field-component
   */
  $field = new BehaviorSubject<string>('');
}
