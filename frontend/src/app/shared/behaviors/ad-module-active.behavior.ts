import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

@Injectable({providedIn: 'root'})
export class AdModuleActiveBehavior {
  $adActive = new BehaviorSubject<boolean>(true);
}
