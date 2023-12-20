import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

@Injectable({providedIn: 'root'})
export class NewAlertBehavior {
  $alertChange = new BehaviorSubject<number>(0);
}
