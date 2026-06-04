import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

@Injectable({providedIn: 'root'})
export class GettingStartedBehavior {
  $init = new BehaviorSubject<boolean>(false);
}
