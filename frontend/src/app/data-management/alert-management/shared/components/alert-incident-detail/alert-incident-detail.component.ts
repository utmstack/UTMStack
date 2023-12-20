import {Component, Input, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {UtmIncidentService} from '../../../../../shared/services/incidents/utm-incident.service';
import {UtmAlertType} from '../../../../../shared/types/alert/utm-alert.type';
import {UtmIncidentType} from '../../../../../shared/types/incident/utm-incident.type';

@Component({
  selector: 'app-alert-incident-detail',
  templateUrl: './alert-incident-detail.component.html',
  styleUrls: ['./alert-incident-detail.component.scss']
})
export class AlertIncidentDetailComponent implements OnInit {
  @Input() alert: UtmAlertType;
  dateCreated: string;
  createdBy: string;
  name: string;
  id: number;
  incident: UtmIncidentType;
  loading = true;

  constructor(private router: Router, private utmIncidentService: UtmIncidentService) {

  }

  ngOnInit() {
    if (this.alert.incidentDetail && this.alert.incidentDetail.incidentId) {
      this.id = Number(this.alert.incidentDetail.incidentId);
      this.createdBy = this.alert.incidentDetail.createdBy;
      this.utmIncidentService.find(this.id).subscribe(response => {
        if (response) {
          this.incident = response.body;
          this.loading = false;
        }
      });
    }
  }

  navigateToIncident() {
    this.router.navigate(['/incident/view'], {queryParams: {incidentId: this.id}});
  }

  isIncident(): boolean {
    return this.alert.isIncident;
  }

  onSuccessMarkAsIncident($event: string) {

  }
}
