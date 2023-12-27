import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

@Injectable({providedIn: 'root'})
export class NavBehavior {
  $nav = new BehaviorSubject<boolean>(null);
}
