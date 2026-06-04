import {Inject, Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

@Injectable({providedIn: 'root'})
export class ModuleRefreshBehavior {
  $moduleChange = new BehaviorSubject<boolean>(null);
}
