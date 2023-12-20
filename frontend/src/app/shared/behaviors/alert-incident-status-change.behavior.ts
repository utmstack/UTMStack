import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

@Injectable({providedIn: 'root'})
export class AlertIncidentStatusChangeBehavior {
  $incidentAlertChange = new BehaviorSubject<number>(0);
}
