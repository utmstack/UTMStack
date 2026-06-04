import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

@Injectable({providedIn: 'root'})
export class VersionUpdateBehavior {
  $versionChange = new BehaviorSubject<boolean>(false);
}
