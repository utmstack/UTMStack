import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

@Injectable({providedIn: 'root'})
export class TreeObjectBehavior {
  $objectId: BehaviorSubject<string> = new BehaviorSubject<string>('');
}
