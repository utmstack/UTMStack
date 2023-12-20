import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

@Injectable({providedIn: 'root'})
export class RestartApiBehavior {
  $restartSignal = new BehaviorSubject<boolean>(null);
}
