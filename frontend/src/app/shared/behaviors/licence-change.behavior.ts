import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

@Injectable({providedIn: 'root'})
export class LicenceChangeBehavior {
  $licenceChange = new BehaviorSubject<boolean>(false);
}
