import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';
import {UtmIncidentType} from '../../../../shared/types/incident/utm-incident.type';


@Injectable({
  providedIn: 'root'
})
export class AlertActionRefreshService {

 private incidentCreatedBehavior = new BehaviorSubject<UtmIncidentType>(null);
  incidentCreated$ = this.incidentCreatedBehavior.asObservable();

  private alertTagRuleCreatedBehavior = new BehaviorSubject<boolean>(false);
  alertTagRuleCreated$ = this.alertTagRuleCreatedBehavior.asObservable();

  constructor() {
  }

  incidentCreated(incident: UtmIncidentType) {
    this.incidentCreatedBehavior.next(incident);
  }

  alertTagRuleCreated(refresh: boolean) {
    this.alertTagRuleCreatedBehavior.next(refresh);
  }

  clearValues() {
    this.incidentCreatedBehavior.next(null);
    this.alertTagRuleCreatedBehavior.next(false);
  }

}
