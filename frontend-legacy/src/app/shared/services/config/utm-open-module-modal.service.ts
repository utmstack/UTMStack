import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {UtmModalActivateType} from '../../types/configuration/utm-modal-activate.type';


@Injectable({
  providedIn: 'root'
})
export class UtmOpenModuleModalService {
  public resourceUrl = SERVER_API_URL + 'api/utm-module-modals';

  constructor(private http: HttpClient) {
  }


  update(tech: UtmModalActivateType): Observable<HttpResponse<UtmModalActivateType>> {
    return this.http.put<UtmModalActivateType>(this.resourceUrl, tech, {observe: 'response'});
  }

  find(id: number): Observable<HttpResponse<UtmModalActivateType>> {
    return this.http.get<UtmModalActivateType>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }

}
