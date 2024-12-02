import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {UtmIncidentType} from '../../types/incident/utm-incident.type';
import {createRequestOption} from '../../util/request-util';
import {ModalConfirmationComponent} from "../../components/utm/util/modal-confirmation/modal-confirmation.component";
import {ModalService} from "../../../core/modal/modal.service";
import {Router} from "@angular/router";


@Injectable({
  providedIn: 'root'
})
export class UtmIncidentService {
  public resourceUrl = SERVER_API_URL + 'api/utm-incidents';

  constructor(private http: HttpClient,
              private modalService: ModalService,
              private router: Router) {
  }


  query(req?: any): Observable<HttpResponse<UtmIncidentType[]>> {
    const options = createRequestOption(req);
    return this.http.get<UtmIncidentType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  update(incident: UtmIncidentType): Observable<HttpResponse<UtmIncidentType>> {
    return this.http.put<UtmIncidentType>(this.resourceUrl, incident, {observe: 'response'});
  }

  changeStatus(incident: UtmIncidentType): Observable<HttpResponse<UtmIncidentType>> {
    return this.http.put<UtmIncidentType>(this.resourceUrl + '/change-status', incident, {observe: 'response'});
  }

  create(incident: any): Observable<HttpResponse<UtmIncidentType>> {
    return this.http.post<UtmIncidentType>(this.resourceUrl, incident, {observe: 'response'});
  }

  addAlerts(incident: any): Observable<HttpResponse<UtmIncidentType>> {
    return this.http.post<UtmIncidentType>(this.resourceUrl + '/add-alerts', incident, {observe: 'response'});
  }

  getUsersAssigned(): Observable<HttpResponse<{ id: number, login: string }[]>> {
    return this.http.get<{ id: number, login: string }[]>(this.resourceUrl + '/users-assigned', {observe: 'response'});
  }

  find(id: number): Observable<HttpResponse<UtmIncidentType>> {
    return this.http.get<UtmIncidentType>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }

  showDialog(msg: string) {
    const incidentModalRef = this.modalService.open(ModalConfirmationComponent, {centered: true});

    incidentModalRef.componentInstance.header = 'Alerts Associated with Another Incident';
    incidentModalRef.componentInstance.message = msg;
    incidentModalRef.componentInstance.confirmBtnText = 'Go to Incidents';
    incidentModalRef.componentInstance.confirmBtnIcon = 'icon-folder';
    incidentModalRef.componentInstance.confirmBtnType = 'default';
    incidentModalRef.result.then(() => {
      this.modalService.close();
      this.router.navigate(['/incident/view']);
    });
  }

}
