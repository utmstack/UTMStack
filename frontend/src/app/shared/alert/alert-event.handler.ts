import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

@Injectable()
export class InverHttpAlertManager {
  public _httpAlert: BehaviorSubject<any> = new BehaviorSubject({});
}
