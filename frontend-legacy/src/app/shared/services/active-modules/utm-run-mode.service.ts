import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {UtmModulesService} from '../../../app-module/shared/services/utm-modules.service';
import {SERVER_API_URL} from '../../../app.constants';

@Injectable({
  providedIn: 'root'
})
export class UtmRunModeService {

  public resourceUrl = SERVER_API_URL + 'api/isLiteMode';

  constructor(private http: HttpClient,
              private moduleService: UtmModulesService) {
  }

  isLiteMode(): Observable<HttpResponse<boolean>> {
    return this.http.get<boolean>(this.resourceUrl, {observe: 'response'});
  }

}
