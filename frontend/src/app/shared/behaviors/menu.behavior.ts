import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

@Injectable({providedIn: 'root'})
export class MenuBehavior {
  $menu = new BehaviorSubject<boolean>(false);
}
